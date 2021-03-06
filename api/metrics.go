package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"minivmm"
)

const promNamespace = "minivmm"

type minivmmExporter struct {
	cpuCores  *prometheus.GaugeVec
	memBytes  *prometheus.GaugeVec
	diskBytes prometheus.Gauge
	numVM     *prometheus.GaugeVec
}

func NewMinivmmExporter() *minivmmExporter {
	return &minivmmExporter{
		cpuCores: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: promNamespace,
				Name:      "cpu_cores",
				Help:      "the summation of usage of cpu cores",
			},
			[]string{"state"},
		),
		memBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: promNamespace,
				Name:      "memory_bytes",
				Help:      "the summation of usage of memory",
			},
			[]string{"state"},
		),
		diskBytes: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: promNamespace,
				Name:      "disk_bytes",
				Help:      "the summation of usage of disk space",
			},
		),
		numVM: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: promNamespace,
				Name:      "vms",
				Help:      "the number of virtual machines",
			},
			[]string{"state"},
		),
	}
}

func (e minivmmExporter) Describe(ch chan<- *prometheus.Desc) {
	e.cpuCores.Describe(ch)
	e.memBytes.Describe(ch)
	ch <- e.diskBytes.Desc()
	e.numVM.Describe(ch)
}

func (e *minivmmExporter) Collect(ch chan<- prometheus.Metric) {
	m, err := minivmm.GetVMMetric()
	if err != nil {
		log.Printf("failed to get metrics; %v", err)
		return
	}

	e.cpuCores.WithLabelValues("all").Set(float64(m.CPUCores))
	e.cpuCores.WithLabelValues("running").Set(float64(m.CPUCoresRunning))
	e.memBytes.WithLabelValues("all").Set(float64(m.MemoryBytes))
	e.memBytes.WithLabelValues("running").Set(float64(m.MemoryBytesRunning))
	e.numVM.WithLabelValues("all").Set(float64(m.NumVM))
	e.numVM.WithLabelValues("running").Set(float64(m.NumVMRunning))

	e.cpuCores.Collect(ch)
	e.memBytes.Collect(ch)
	ch <- prometheus.MustNewConstMetric(e.diskBytes.Desc(), prometheus.GaugeValue, float64(m.DiskBytes))
	e.numVM.Collect(ch)
}

// HandleJsonMetrics handles json metrics request.
func HandleJsonMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	vmMetric, err := minivmm.GetVMMetric()
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	sysMetric, err := minivmm.GetSysMetric()
	if err != nil {
		writeInternalServerError(err, w)
		return
	}

	m := struct {
		VM  *minivmm.VMMetric  `json:"vm"`
		Sys *minivmm.SysMetric `json:"sys"`
	}{vmMetric, sysMetric}
	b, _ := json.Marshal(m)
	w.Write(b)
}

// GetMetricsHandler returns the prometheus metrics handler.
func GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}

// InitMetricsHandler initializes the prometheus metrics handler.
func InitMetricsHandler() {
	prometheus.MustRegister(NewMinivmmExporter())
}
