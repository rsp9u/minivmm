package minivmm

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var (
	vsizeRegexp = regexp.MustCompile(`virtual size: ([0-9.]+) (([A-Z]i)?B) \(.*\)`)
	dsizeRegexp = regexp.MustCompile(`disk size: ([0-9.]+) (([A-Z]i)?B)`)
)

// ListBaseImages returns a list of base images from file system.
func ListBaseImages() []string {
	d, _ := os.Open(ImageDir)
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

	b, _ := filepath.Abs(filepath.Join(ImageDir, base))
	p, _ := filepath.Abs(filepath.Join(dstDir, name+".qcow2"))
	o := fmt.Sprintf("backing_file=%s,backing_fmt=qcow2", b)
	o2 := "cluster_size=2M"
	f := "qcow2"
	params := [][]string{{"qemu-img", "create", "-f", f, "-o", o, "-o", o2, p, size}}

	log.Println("Creating image: ", params)
	err = Execs(params)
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
	params := [][]string{{"qemu-img", "info", path}}
	stdouts, err := ExecsStdout(params)
	if err != nil {
		return err
	}

	m := vsizeRegexp.FindAllStringSubmatch(stdouts[0], -1)
	if len(m) < 1 || len(m[0]) < 3 {
		return errors.New("invalid command result: not found expected virtual size line")
	}
	vsize, err := convertSISize(m[0][1], m[0][2])
	if err != nil {
		return err
	}

	m = dsizeRegexp.FindAllStringSubmatch(stdouts[0], -1)
	if len(m) < 1 || len(m[0]) < 3 {
		return errors.New("invalid command result: not found expected disk size line")
	}
	dsize, _ := convertSISize(m[0][1], m[0][2])

	if vsize < dsize {
		return errors.New("the given virtual disk size is smaller than base size")
	}

	return nil
}

func convertSISize(value, unit string) (int, error) {
	s, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.Wrap(err, "failed to convert disk size value: "+value)
	}

	if unit == "B" {
		return s, nil
	}
	if unit == "KiB" {
		return s * 1024, nil
	}
	if unit == "MiB" {
		return s * 1024 * 1024, nil
	}
	if unit == "GiB" {
		return s * 1024 * 1024 * 1024, nil
	}
	if unit == "TiB" {
		return s * 1024 * 1024 * 1024 * 1024, nil
	}

	return 0, errors.New("not supported unit: " + unit)
}
