package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"strings"
	"time"

	"github.com/md14454/gosensors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	fanspeed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "fan_speed_rpm",
		Help:      "fan speed (rotations per minute).",
	}, []string{"fantype", "chip", "adaptor"})

	temperature = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "temperature_celsius",
		Help:      "temperature in celsius",
	}, []string{"temptype", "chip", "adaptor"})
)

func init() {
	prometheus.MustRegister(fanspeed)
	prometheus.MustRegister(temperature)
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9255", "Address on which to expose metrics and web interface.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
	)
	flag.Parse()

	go collect()

	http.Handle(*metricsPath, prometheus.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Sensor Exporter</title></head>
			<body>
			<h1>Sensor Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})
	http.ListenAndServe(*listenAddress, nil)
}

func collect() {
	gosensors.Init()
	defer gosensors.Cleanup()
	for {
		for _, chip := range gosensors.GetDetectedChips() {
			chipName := chip.String()
			adaptorName := chip.AdapterName()
			for _, feature := range chip.GetFeatures() {
				if strings.HasPrefix(feature.Name, "fan") {
					fanspeed.WithLabelValues(feature.GetLabel(), chipName, adaptorName).Set(feature.GetValue())
				} else if strings.HasPrefix(feature.Name, "temp") {
					temperature.WithLabelValues(feature.GetLabel(), chipName, adaptorName).Set(feature.GetValue())
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
}
