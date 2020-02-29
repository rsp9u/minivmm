package minivmm

import (
	"log"
	"strconv"
)

// Metric is the metrics of resrouces that minivmm manages. It's exported for prometheus.
type Metric struct {
	CPUCores           int `json:"minivmm_cpu_cores"`
	CPUCoresRunning    int `json:"minivmm_cpu_cores_running"`
	MemoryBytes        int `json:"minivmm_memory_bytes"`
	MemoryBytesRunning int `json:"minivmm_memory_bytes_running"`
	DiskBytes          int `json:"minivmm_disk_bytes"`
	NumVM              int `json:"minivmm_vms"`
	NumVMRunning       int `json:"minivmm_vms_running"`
}

func GetMetric() (*Metric, error) {
	var m Metric

	vms, err := ListVMs()
	if err != nil {
		return nil, err
	}

	m.NumVM = len(vms)
	for _, vm := range vms {
		cpuStr := vm.CPU
		memStr, _ := convertSIPrefixedValue(vm.Memory, "")
		diskStr, _ := convertSIPrefixedValue(vm.Disk, "")

		cpu, err := strconv.Atoi(cpuStr)
		if err != nil {
			log.Printf("failed to parse cpu info, %v\n", err)
			continue
		}
		mem, err := strconv.Atoi(memStr)
		if err != nil {
			log.Printf("failed to parse memory info, %v\n", err)
			continue
		}
		disk, err := strconv.Atoi(diskStr)
		if err != nil {
			log.Printf("failed to parse disk info, %v\n", err)
			continue
		}

		m.CPUCores += cpu
		m.MemoryBytes += mem
		m.DiskBytes += disk
		if vm.Status == "running" {
			m.NumVMRunning += 1
			m.CPUCoresRunning += cpu
			m.MemoryBytesRunning += mem
		}
	}

	return &m, nil
}
