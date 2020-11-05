package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"minivmm"
)

var (
	updateVMAPI    = regexp.MustCompile(`^/api/v1/vms/[^/]+$`)
	extraVolumeAPI = regexp.MustCompile(`^/api/v1/vms/[^/]+/volumes.*$`)
)

type vm struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	Owner        string        `json:"owner"`
	Hypervisor   string        `json:"hypervisor"`
	Image        string        `json:"image"`
	IP           string        `json:"ip"`
	CPU          string        `json:"cpu"`
	Memory       string        `json:"memory"`
	Disk         string        `json:"disk"`
	Tag          string        `json:"tag"`
	Lock         string        `json:"lock"`
	UserData     string        `json:"user_data"`
	ExtraVolumes []extraVolume `json:"extra_volumes"`
}

type extraVolume struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// HandleVMs handles virtual machine resource request.
func HandleVMs(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && extraVolumeAPI.MatchString(r.URL.String()) {
		CreateVolume(w, r)
		return
	}
	if r.Method == http.MethodDelete && extraVolumeAPI.MatchString(r.URL.String()) {
		DeleteVolume(w, r)
		return
	}

	if r.Method == http.MethodGet {
		ListVMs(w, r)
		return
	}
	if r.Method == http.MethodPost {
		CreateVM(w, r)
		return
	}
	if r.Method == http.MethodPatch && updateVMAPI.MatchString(r.URL.String()) {
		UpdateVM(w, r)
		return
	}
	if r.Method == http.MethodDelete {
		RemoveVM(w, r)
		return
	}
	w.WriteHeader(http.StatusMethodNotAllowed)
}

// ListVMs returns a list of VMs.
func ListVMs(w http.ResponseWriter, r *http.Request) {
	vmMetaData, err := minivmm.ListVMs()
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	// convert metadata to api-vm struct
	hostname, _ := os.Hostname()
	vms := []*vm{}
	for _, metaData := range vmMetaData {
		if metaData.Owner != minivmm.GetUserName(r) {
			continue
		}
		ev := []extraVolume{}
		if metaData.ExtraVolumes != nil {
			for _, vol := range metaData.ExtraVolumes {
				ev = append(ev, extraVolume{vol.Name, vol.Size})
			}
		}
		vm := vm{
			Name:         metaData.Name,
			Status:       metaData.Status,
			Owner:        metaData.Owner,
			Hypervisor:   hostname,
			Image:        metaData.Image,
			IP:           metaData.IPAddress,
			CPU:          metaData.CPU,
			Memory:       metaData.Memory,
			Disk:         metaData.Disk,
			Lock:         strconv.FormatBool(metaData.Lock),
			Tag:          metaData.Tag,
			ExtraVolumes: ev,
		}
		vms = append(vms, &vm)
	}
	ret := map[string][]*vm{"vms": vms}
	b, _ := json.Marshal(ret)
	w.Write(b)
}

// CreateVM creates VM and writes its metadata.
func CreateVM(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	v, err := parseVMFromJson(r.Body)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	metaData, err := minivmm.CreateVM(v.Name, minivmm.GetUserName(r), v.Image, v.CPU, v.Memory, v.Disk, v.UserData, v.Tag)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	b, _ := json.Marshal(metaData)
	w.Write(b)
}

// UpdateVM update VM's state.
func UpdateVM(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.String(), "/")
	vmName := paths[len(paths)-1]

	err := restrictVMOperationByOwner(w, r, vmName)
	if err != nil {
		return
	}

	defer r.Body.Close()
	v, err := parseVMFromJson(r.Body)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	if v.Status != "" {
		if v.Status == "start" {
			_, err = minivmm.StartVM(vmName)
		} else if v.Status == "stop" {
			err = minivmm.StopVM(vmName)
		}
		if err != nil {
			writeInternalServerError(err, w)
			return
		}
	}

	if v.CPU != "" || v.Memory != "" || v.Disk != "" {
		metaData, err := resizeVM(vmName, v)
		if err != nil {
			writeInternalServerError(err, w)
			return
		}

		b, _ := json.Marshal(metaData)
		w.Write(b)
	}

	if v.Lock != "" {
		var metaData *minivmm.VMMetaData
		if v.Lock == "true" {
			metaData, err = minivmm.LockVM(vmName)
		} else {
			metaData, err = minivmm.UnlockVM(vmName)
		}

		if err != nil {
			writeInternalServerError(err, w)
			return
		}

		b, _ := json.Marshal(metaData)
		w.Write(b)
	}
}

// RemoveVM remove VM
func RemoveVM(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.String(), "/")
	vmName := paths[len(paths)-1]

	err := restrictVMOperationByOwner(w, r, vmName)
	if err != nil {
		return
	}

	err = minivmm.RemoveVM(vmName)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func resizeVM(vmName string, v *vm) (*minivmm.VMMetaData, error) {
	err := minivmm.StopVM(vmName)
	if err != nil {
		return nil, err
	}

	metaData, err := minivmm.ResizeVM(vmName, v.CPU, v.Memory, v.Disk)
	if err != nil {
		return nil, err
	}

	if v.Disk != "" {
		vmDataDir := filepath.Join(minivmm.C.VMDir, vmName)
		err := minivmm.ResizeImage(vmName, v.Disk, vmDataDir)
		if err != nil {
			return nil, err
		}
	}

	_, err = minivmm.StartVM(vmName)
	if err != nil {
		return nil, err
	}

	return metaData, nil
}

func restrictVMOperationByOwner(w http.ResponseWriter, r *http.Request, vmName string) error {
	metaData, err := minivmm.GetVM(vmName)
	if err != nil {
		writeInternalServerError(err, w)
		return err
	}

	if metaData.Owner != minivmm.GetUserName(r) {
		writeForbidden(w)
		return fmt.Errorf("forbidden")
	}

	return nil
}

// CreateVolume adds a new extra volume to the VM.
func CreateVolume(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.String(), "/")
	vmName := paths[len(paths)-2]

	err := restrictVMOperationByOwner(w, r, vmName)
	if err != nil {
		return
	}

	defer r.Body.Close()
	volSize, err := parseExtraVolumeSize(r.Body)
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	metaData, err := minivmm.AddVolume(vmName, volSize)

	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	b, _ := json.Marshal(metaData)
	w.Write(b)
}

// DeleteVolume removes an extra volume from the VM.
func DeleteVolume(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.String(), "/")
	volName := paths[len(paths)-1]
	vmName := paths[len(paths)-3]

	err := restrictVMOperationByOwner(w, r, vmName)
	if err != nil {
		return
	}

	metaData, err := minivmm.RemoveVolume(vmName, volName)

	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	b, _ := json.Marshal(metaData)
	w.Write(b)
}

func parseVMFromJson(r io.ReadCloser) (*vm, error) {
	// read and parse json
	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	var v vm
	err := json.Unmarshal(buf.Bytes(), &v)
	if err != nil {
		return nil, err
	}
	fmt.Printf("%v\n", v)

	// sanitize
	err = rfc1035Label(v.Name)
	if err != nil {
		return nil, err
	}
	err = numeric(v.CPU, "cpu")
	if err != nil {
		return nil, err
	}
	err = alphanumeric(v.Memory, "memory")
	if err != nil {
		return nil, err
	}
	err = alphanumeric(v.Disk, "disk")
	if err != nil {
		return nil, err
	}

	return &v, nil
}

func parseExtraVolumeSize(r io.ReadCloser) (string, error) {
	// read and parse json
	buf := new(bytes.Buffer)
	io.Copy(buf, r)

	var ev extraVolume
	err := json.Unmarshal(buf.Bytes(), &ev)
	if err != nil {
		return "", err
	}
	fmt.Printf("%v\n", ev)

	return ev.Size, nil
}

func rfc1035Label(label string) error {
	errmsg := "name contains alphanumeric or hyphen, and must begin with a letter"
	for i, c := range label {
		if c > unicode.MaxASCII {
			return fmt.Errorf(errmsg)
		}
		if i == 0 && !unicode.IsLetter(c) {
			return fmt.Errorf(errmsg)
		}
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' {
			return fmt.Errorf(errmsg)
		}
	}
	return nil
}

func alphanumeric(s, kind string) error {
	errmsg := fmt.Sprintf("%s contains alphanumeric characters", kind)
	for i, c := range s {
		if c > unicode.MaxASCII {
			return fmt.Errorf(errmsg)
		}
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
			return fmt.Errorf(errmsg)
		}
	}
}

func numeric(s, kind string) error {
	errmsg := fmt.Sprintf("%s contains numeric characters", kind)
	for i, c := range s {
		if c > unicode.MaxASCII {
			return fmt.Errorf(errmsg)
		}
		if !unicode.IsDigit(c) {
			return fmt.Errorf(errmsg)
		}
	}
}
