package resource

import (
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"minivmm"
)

// SysMetric is the metrics of system resrouces.
type SysMetric struct {
	CPUCores    int    `json:"minivmm_sys_cpu_cores"`
	MemoryBytes uint64 `json:"minivmm_sys_memory_bytes"`
	DiskBytes   uint64 `json:"minivmm_sys_disk_bytes"`
}

// GetSysMetric returns the system resource metrics.
func GetSysMetric() (*SysMetric, error) {
	var m SysMetric

	cpu, err := cpu.Get()
	if err != nil {
		return nil, err
	}
	m.CPUCores = cpu.CPUCount

	memory, err := memory.Get()
	if err != nil {
		return nil, err
	}
	m.MemoryBytes = memory.Total

	m.DiskBytes = getDiskSize(minivmm.C.Dir)

	return &m, nil
}
