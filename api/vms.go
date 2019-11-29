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
	"strings"

	"minivmm"
)

var (
	updateVMAPI = regexp.MustCompile(`^/api/v1/vms/[^/]+$`)
)

type vm struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Owner      string `json:"owner"`
	Hypervisor string `json:"hypervisor"`
	Image      string `json:"image"`
	IP         string `json:"ip"`
	CPU        string `json:"cpu"`
	Memory     string `json:"memory"`
	Disk       string `json:"disk"`
	Tag        string `json:"tag"`
	UserData   string `json:"user_data"`
}

// HandleVMs handles virtual machine resource request.
func HandleVMs(w http.ResponseWriter, r *http.Request) {
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
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}

	// convert metadata to api-vm struct
	hostname, _ := os.Hostname()
	vms := []*vm{}
	for _, metaData := range vmMetaData {
		if metaData.Owner != minivmm.GetUserName(r) {
			continue
		}
		vm := vm{
			Name:       metaData.Name,
			Status:     metaData.Status,
			Owner:      metaData.Owner,
			Hypervisor: hostname,
			Image:      metaData.Image,
			IP:         metaData.IPAddress,
			CPU:        metaData.CPU,
			Memory:     metaData.Memory,
			Disk:       metaData.Disk,
			Tag:        metaData.Tag,
		}
		vms = append(vms, &vm)
	}
	ret := map[string][]*vm{"vms": vms}
	b, _ := json.Marshal(ret)
	w.Write(b)
}

// CreateVM creates VM and writes its metadata.
func CreateVM(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)

	var v vm
	json.Unmarshal(buf.Bytes(), &v)
	fmt.Printf("%v\n", v)

	metaData, err := minivmm.CreateVM(v.Name, minivmm.GetUserName(r), v.Image, v.CPU, v.Memory, v.Disk, v.UserData, v.Tag)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
		return
	}

	b, _ := json.Marshal(metaData)
	w.Write(b)
}

// UpdateVM update VM's state.
func UpdateVM(w http.ResponseWriter, r *http.Request) {
	paths := strings.Split(r.URL.String(), "/")
	vmName := paths[len(paths)-1]

	body := r.Body
	defer body.Close()

	buf := new(bytes.Buffer)
	io.Copy(buf, body)

	var v vm
	json.Unmarshal(buf.Bytes(), &v)
	fmt.Printf("%v\n", v)

	var err error = nil

	if v.Status != "" {
		if v.Status == "start" {
			_, err = minivmm.StartVM(vmName)
		} else if v.Status == "stop" {
			err = minivmm.StopVM(vmName)
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ret := map[string]string{"error": err.Error()}
			b, _ := json.Marshal(ret)
			w.Write(b)
			return
		}
	}

	if v.CPU != "" || v.Memory != "" || v.Disk != "" {
		metaData, err := resizeVM(vmName, &v)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			ret := map[string]string{"error": err.Error()}
			b, _ := json.Marshal(ret)
			w.Write(b)
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
	err := minivmm.RemoveVM(vmName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		ret := map[string]string{"error": err.Error()}
		b, _ := json.Marshal(ret)
		w.Write(b)
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
		vmDataDir := filepath.Join(minivmm.VMDir, vmName)
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
