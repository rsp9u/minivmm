package minivmm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// ListBaseImages returns a list of base images from file system.
func ListBaseImages() []string {
	d, _ := os.Open(C.ImageDir)
	files, _ := d.Readdir(0)

	names := []string{}
	for _, f := range files {
		names = append(names, f.Name())
	}

	return names
}

// CreateImage creates a new image with backing file. If created image virtual size is lesser than disk size, this will return error, but created image file won't be removed.
func CreateImage(name, size, base, dstDir string) (string, error) {
	err := os.MkdirAll(dstDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	params := []string{"qemu-img", "create", "-f", "qcow2", "-o", "cluster_size=2M"}
	if base != "" {
		b, _ := filepath.Abs(filepath.Join(C.ImageDir, base))
		o := fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", b)
		params = append(params, "-o")
		params = append(params, o)
	}
	p, _ := filepath.Abs(filepath.Join(dstDir, name+".qcow2"))
	params = append(params, p)
	params = append(params, size)

	log.Println("Creating image: ", params)
	err = Execs([][]string{params})
	if err != nil {
		return "", err
	}

	err = checkImageSize(p)
	if err != nil {
		return "", err
	}

	return p, nil
}

// ResizeImage resizes the image size.
func ResizeImage(name, size, dstDir string) error {
	p, _ := filepath.Abs(filepath.Join(dstDir, name+".qcow2"))
	c := [][]string{{"qemu-img", "resize", p, size}}
	err := Execs(c)
	if err != nil {
		return err
	}
	return nil
}

func checkImageSize(path string) error {
	params := [][]string{{"qemu-img", "info", "--output", "json", path}}
	stdouts, err := ExecsStdout(params)
	if err != nil {
		return err
	}

	var imageInfo struct {
		VirtualSize string `json:"virtual-size"`
		ActualSize  string `json:"actual-size"`
	}
	json.Unmarshal([]byte(stdouts[0]), &imageInfo)

	if imageInfo.VirtualSize < imageInfo.ActualSize {
		return errors.New("the given virtual disk size is smaller than base size")
	}

	return nil
}
