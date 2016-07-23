package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
	"time"

	"github.com/ncabatoff/gosensors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	fanspeed = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "fan_speed_rpm",
		Help:      "fan speed (rotations per minute).",
	}, []string{"fantype", "chip", "adaptor"})

	voltages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "voltage_volts",
		Help:      "voltage in volts",
	}, []string{"intype", "chip", "adaptor"})

	powers = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "power_watts",
		Help:      "power in watts",
	}, []string{"powertype", "chip", "adaptor"})

	temperature = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "lm",
		Name:      "temperature_celsius",
		Help:      "temperature in celsius",
	}, []string{"temptype", "chip", "adaptor"})

	hddtemperature = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "sensor",
		Subsystem: "hddsmart",
		Name:      "temperature_celsius",
		Help:      "temperature in celsius",
	}, []string{"device", "id"})
)

func init() {
	prometheus.MustRegister(fanspeed)
	prometheus.MustRegister(voltages)
	prometheus.MustRegister(powers)
	prometheus.MustRegister(temperature)
	prometheus.MustRegister(hddtemperature)
}

func main() {
	var (
		listenAddress  = flag.String("web.listen-address", ":9255", "Address on which to expose metrics and web interface.")
		metricsPath    = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		hddtempAddress = flag.String("hddtemp-address", "localhost:7634", "Address to fetch hdd metrics from.")
	)
	flag.Parse()

	go collectLm()
	go collectHdd(*hddtempAddress)

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

func collectLm() {
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
				} else if strings.HasPrefix(feature.Name, "in") {
					voltages.WithLabelValues(feature.GetLabel(), chipName, adaptorName).Set(feature.GetValue())
				} else if strings.HasPrefix(feature.Name, "power") {
					powers.WithLabelValues(feature.GetLabel(), chipName, adaptorName).Set(feature.GetValue())
				}
			}

		}
		time.Sleep(1 * time.Second)
	}
}

func collectHdd(address string) {
	for {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Printf("Error connecting to hddtemp address '%s': %v", address, err)
		} else {
			var buf bytes.Buffer
			_, err := io.Copy(&buf, conn)
			if err != nil {
				log.Printf("Error reading from hddtemp socket: %v", err)
			} else {
				parseHddTemps(buf.String())
			}
			if err := conn.Close(); err != nil {
				log.Printf("Error closing hddtemp socket: %v", err)
			}
		}
		time.Sleep(1 * time.Second)
	}
}

func parseHddTemps(s string) {
	if len(s) < 1 || s[0] != '|' {
		log.Printf("Error parsing output from hddtemp: %s", s)
	}
	for _, item := range strings.Split(s[1:len(s)-1], "||") {
		pieces := strings.Split(item, "|")
		if len(pieces) != 4 {
			log.Printf("Error parsing item from hddtemp, expected 4 tokens: %s", item)
		} else if pieces[3] != "C" {
			log.Printf("Error parsing item from hddtemp, I only speak Celsius", item)
		} else {
			dev, id, temp := pieces[0], pieces[1], pieces[2]
			ftemp, err := strconv.ParseFloat(temp, 64)
			if err != nil {
				log.Printf("Error parsing temperature as float: %s", temp)
			} else {
				hddtemperature.WithLabelValues(dev, id).Set(ftemp)
				// TODO handle unit F
			}
		}
	}
}
