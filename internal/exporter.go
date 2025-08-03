package internal

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	log "github.com/sirupsen/logrus"
)

const namespace = "unifi_ap_"

// Try to match unpoller metrics and labels as much as possible
// See https://github.com/unpoller/unpoller/tree/master/pkg/promunifi
type deviceMetrics struct {
	info         *prometheus.Desc
	uptime       *prometheus.Desc
	totalTxBytes *prometheus.Desc
	totalRxBytes *prometheus.Desc
	loadAvg1     *prometheus.Desc
	loadAvg5     *prometheus.Desc
	loadAvg15    *prometheus.Desc
	memUsed      *prometheus.Desc
	memTotal     *prometheus.Desc
	memBuffer    *prometheus.Desc
	cpu          *prometheus.Desc
	mem          *prometheus.Desc
}

type radioMetrics struct {
	currentAntennaGain *prometheus.Desc
	maxTxpower         *prometheus.Desc
	minTxpower         *prometheus.Desc
}

type vapMetrics struct {
	rxBytes   *prometheus.Desc
	rxDropped *prometheus.Desc
	rxErrors  *prometheus.Desc
	txBytes   *prometheus.Desc
	txDropped *prometheus.Desc
	txErrors  *prometheus.Desc
	txPower   *prometheus.Desc
	txRetries *prometheus.Desc
	txSuccess *prometheus.Desc
	txTotal   *prometheus.Desc
}

type stationMetrics struct {
	txBytes *prometheus.Desc
	rxBytes *prometheus.Desc
	noise   *prometheus.Desc
	signal  *prometheus.Desc
}

type rogueMetrics struct {
	channel   *prometheus.Desc
	frequency *prometheus.Desc
	noise     *prometheus.Desc
	signal    *prometheus.Desc
}

type Exporter struct {
	collector Collector
	device    deviceMetrics
	radio     radioMetrics
	vap       vapMetrics
	station   stationMetrics
	rogue     rogueMetrics
}

func NewExporter(collector Collector) *Exporter {
	var deviceInfoLabels = []string{"ip", "mac", "model", "model_name", "name", "serial", "version"}
	var deviceLabels = []string{"name", "model"}
	var radioLabels = []string{"name", "radio", "radio_name"}
	var vapLabels = []string{"name", "vap_name", "bssid", "radio", "radio_name", "essid", "usage"}
	var stationLabels = []string{"name", "vap_name", "hostname", "mac"}
	var rogueLabels = []string{"name", "radio", "bssid", "essid", "security"}

	var DeviceMetrics = deviceMetrics{
		info:         prometheus.NewDesc(namespace+"info", "Device Information", deviceInfoLabels, nil),
		uptime:       prometheus.NewDesc(namespace+"uptime_seconds", "Device Uptime", deviceLabels, nil),
		totalTxBytes: prometheus.NewDesc(namespace+"transmit_bytes_total", "Total Transmitted Bytes", deviceLabels, nil),
		totalRxBytes: prometheus.NewDesc(namespace+"receive_bytes_total", "Total Received Bytes", deviceLabels, nil),
		loadAvg1:     prometheus.NewDesc(namespace+"load_average_1", "System Load Average 1 Minute", deviceLabels, nil),
		loadAvg5:     prometheus.NewDesc(namespace+"load_average_5", "System Load Average 5 Minutes", deviceLabels, nil),
		loadAvg15:    prometheus.NewDesc(namespace+"load_average_15", "System Load Average 15 Minutes", deviceLabels, nil),
		memUsed:      prometheus.NewDesc(namespace+"memory_used_bytes", "System Memory Used", deviceLabels, nil),
		memTotal:     prometheus.NewDesc(namespace+"memory_installed_bytes", "System Installed Memory", deviceLabels, nil),
		memBuffer:    prometheus.NewDesc(namespace+"memory_buffer_bytes", "System Memory Buffer", deviceLabels, nil),
		cpu:          prometheus.NewDesc(namespace+"cpu_utilization_ratio", "System CPU % Utilized", deviceLabels, nil),
		mem:          prometheus.NewDesc(namespace+"memory_utilization_ratio", "System Memory % Utilized", deviceLabels, nil),
	}
	var RadioMetrics = radioMetrics{
		currentAntennaGain: prometheus.NewDesc(namespace+"radio_current_antenna_gain", "Radio Current Antenna Gain", radioLabels, nil),
		maxTxpower:         prometheus.NewDesc(namespace+"radio_max_transmit_power", "Radio Maximum Transmit Power", radioLabels, nil),
		minTxpower:         prometheus.NewDesc(namespace+"radio_min_transmit_power", "Radio Minimum Transmit Power", radioLabels, nil),
	}
	var VapMetrics = vapMetrics{
		rxBytes:   prometheus.NewDesc(namespace+"vap_receive_bytes_total", "VAP Bytes Received", vapLabels, nil),
		rxDropped: prometheus.NewDesc(namespace+"vap_receive_dropped_total", "VAP Dropped Received", vapLabels, nil),
		rxErrors:  prometheus.NewDesc(namespace+"vap_receive_errors_total", "VAP Errors Received", vapLabels, nil),
		txBytes:   prometheus.NewDesc(namespace+"vap_transmit_bytes_total", "VAP Bytes Transmitted", vapLabels, nil),
		txDropped: prometheus.NewDesc(namespace+"vap_transmit_dropped_total", "VAP Dropped Transmitted", vapLabels, nil),
		txErrors:  prometheus.NewDesc(namespace+"vap_transmit_errors_total", "VAP Errors Transmitted", vapLabels, nil),
		txPower:   prometheus.NewDesc(namespace+"vap_transmit_power", "VAP Transmit Power", vapLabels, nil),
		txRetries: prometheus.NewDesc(namespace+"vap_transmit_retries_total", "VAP Retries Transmitted", vapLabels, nil),
		txSuccess: prometheus.NewDesc(namespace+"vap_transmit_success_total", "VAP Success Transmits", vapLabels, nil),
		txTotal:   prometheus.NewDesc(namespace+"vap_transmit_total", "VAP Transmit Total", vapLabels, nil),
	}
	var StationMetrics = stationMetrics{
		txBytes: prometheus.NewDesc(namespace+"station_transmit_bytes_total", "Station Bytes Transmitted", stationLabels, nil),
		rxBytes: prometheus.NewDesc(namespace+"station_receive_bytes_total", "Station Bytes Received", stationLabels, nil),
		noise:   prometheus.NewDesc(namespace+"station_noise", "Station Noise", stationLabels, nil),
		signal:  prometheus.NewDesc(namespace+"station_signal", "Station Signal", stationLabels, nil),
	}
	var RogueMetrics = rogueMetrics{
		channel:   prometheus.NewDesc(namespace+"rogueap_channel", "RogueAP Channel", rogueLabels, nil),
		frequency: prometheus.NewDesc(namespace+"rogueap_frequency", "RogueAP Frequency", rogueLabels, nil),
		noise:     prometheus.NewDesc(namespace+"rogueap_noise", "RogueAP Noise", rogueLabels, nil),
		signal:    prometheus.NewDesc(namespace+"rogueap_signal", "RogueAP Signal", rogueLabels, nil),
	}

	return &Exporter{
		collector: collector,
		device:    DeviceMetrics,
		radio:     RadioMetrics,
		vap:       VapMetrics,
		station:   StationMetrics,
		rogue:     RogueMetrics,
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// Device metrics
	ch <- e.device.info
	ch <- e.device.uptime
	ch <- e.device.totalTxBytes
	ch <- e.device.totalRxBytes
	ch <- e.device.loadAvg1
	ch <- e.device.loadAvg5
	ch <- e.device.loadAvg15
	ch <- e.device.memUsed
	ch <- e.device.memTotal
	ch <- e.device.memBuffer
	ch <- e.device.cpu
	ch <- e.device.mem
	// Radio metrics
	ch <- e.radio.currentAntennaGain
	ch <- e.radio.maxTxpower
	ch <- e.radio.minTxpower
	// Virtual Accesspoint metrics
	ch <- e.vap.rxBytes
	ch <- e.vap.rxDropped
	ch <- e.vap.rxErrors
	ch <- e.vap.txBytes
	ch <- e.vap.txDropped
	ch <- e.vap.txErrors
	ch <- e.vap.txPower
	ch <- e.vap.txRetries
	ch <- e.vap.txSuccess
	ch <- e.vap.txTotal
	// Station (client) metrics
	ch <- e.station.rxBytes
	ch <- e.station.txBytes
	ch <- e.station.noise
	ch <- e.station.signal
	// Rogue AP (others) metrics
	ch <- e.rogue.channel
	ch <- e.rogue.frequency
	ch <- e.rogue.noise
	ch <- e.rogue.signal
}

func (e *Exporter) Run() {
	prometheus.MustRegister(e)
	http.Handle("/metrics", promhttp.Handler())
	log.Info("listening for requests on port ", e.collector.config.Global.ListenPort)
	log.Fatal(http.ListenAndServe(fmt.Sprint(":", e.collector.config.Global.ListenPort), nil))
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	accessPointInfos, err := e.collector.Collect()
	if err != nil {
		log.Errorf("collect failed: %s", err)
		return
	}

	for _, accessPointInfo := range *accessPointInfos {
		// Device info
		ch <- prometheus.MustNewConstMetric(e.device.info, prometheus.GaugeValue, accessPointInfo.Value,
			accessPointInfo.IP, accessPointInfo.Mac, accessPointInfo.Model, accessPointInfo.ModelName,
			accessPointInfo.Name, accessPointInfo.Serial, accessPointInfo.Version)
		ch <- prometheus.MustNewConstMetric(e.device.uptime, prometheus.GaugeValue, float64(accessPointInfo.Uptime),
			accessPointInfo.Name, accessPointInfo.Model)
		// Bytes sent and received
		var txTotal = int64(0)
		var rxTotal = int64(0)
		for _, i := range accessPointInfo.InterfaceTable {
			if i.Up {
				txTotal += i.TxBytes
				rxTotal += i.RxBytes
			}
		}
		ch <- prometheus.MustNewConstMetric(e.device.totalTxBytes, prometheus.CounterValue, float64(txTotal),
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.totalRxBytes, prometheus.CounterValue, float64(rxTotal),
			accessPointInfo.Name, accessPointInfo.Model)

		// CPU and memory
		ch <- prometheus.MustNewConstMetric(e.device.loadAvg1, prometheus.GaugeValue, accessPointInfo.SysStats.LoadAvg1,
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.loadAvg5, prometheus.GaugeValue, accessPointInfo.SysStats.LoadAvg5,
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.loadAvg15, prometheus.GaugeValue, accessPointInfo.SysStats.LoadAvg15,
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.memUsed, prometheus.GaugeValue, float64(accessPointInfo.SysStats.MemUsed),
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.memTotal, prometheus.GaugeValue, float64(accessPointInfo.SysStats.MemTotal),
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.memBuffer, prometheus.GaugeValue, float64(accessPointInfo.SysStats.MemBuffer),
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.cpu, prometheus.GaugeValue, accessPointInfo.SystemStats.CPU,
			accessPointInfo.Name, accessPointInfo.Model)
		ch <- prometheus.MustNewConstMetric(e.device.mem, prometheus.GaugeValue, accessPointInfo.SystemStats.Mem,
			accessPointInfo.Name, accessPointInfo.Model)

		// Radio
		for _, radio := range accessPointInfo.RadioTable {
			ch <- prometheus.MustNewConstMetric(e.radio.currentAntennaGain, prometheus.GaugeValue, float64(radio.CurrentAntennaGain),
				accessPointInfo.Name, radio.Radio, radio.RadioName)
			ch <- prometheus.MustNewConstMetric(e.radio.maxTxpower, prometheus.GaugeValue, float64(radio.MaxTxpower),
				accessPointInfo.Name, radio.Radio, radio.RadioName)
			ch <- prometheus.MustNewConstMetric(e.radio.minTxpower, prometheus.GaugeValue, float64(radio.MinTxpower),
				accessPointInfo.Name, radio.Radio, radio.RadioName)

			// Rogue AP (others)
			for _, rogue := range radio.ScanTable {
				ch <- prometheus.MustNewConstMetric(e.rogue.frequency, prometheus.GaugeValue, float64(rogue.Frequency),
					accessPointInfo.Name, radio.Radio, rogue.BSSID, rogue.ESSID, rogue.Security)
				ch <- prometheus.MustNewConstMetric(e.rogue.channel, prometheus.GaugeValue, float64(rogue.Channel),
					accessPointInfo.Name, radio.Radio, rogue.BSSID, rogue.ESSID, rogue.Security)
				ch <- prometheus.MustNewConstMetric(e.rogue.noise, prometheus.GaugeValue, float64(rogue.Noise),
					accessPointInfo.Name, radio.Radio, rogue.BSSID, rogue.ESSID, rogue.Security)
				ch <- prometheus.MustNewConstMetric(e.rogue.signal, prometheus.GaugeValue, float64(rogue.Signal),
					accessPointInfo.Name, radio.Radio, rogue.BSSID, rogue.ESSID, rogue.Security)
			}
		}

		// Virtual Accesspoint
		for _, vap := range accessPointInfo.VAPTable {
			ch <- prometheus.MustNewConstMetric(e.vap.rxBytes, prometheus.CounterValue, float64(vap.RxBytes),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.rxDropped, prometheus.CounterValue, float64(vap.RxDropped),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.rxErrors, prometheus.CounterValue, float64(vap.RxErrors),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txBytes, prometheus.CounterValue, float64(vap.TxBytes),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txDropped, prometheus.CounterValue, float64(vap.TxDropped),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txErrors, prometheus.CounterValue, float64(vap.TxErrors),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txPower, prometheus.GaugeValue, float64(vap.TxPower),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txRetries, prometheus.CounterValue, float64(vap.TxRetries),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txSuccess, prometheus.CounterValue, float64(vap.TxSuccess),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)
			ch <- prometheus.MustNewConstMetric(e.vap.txTotal, prometheus.CounterValue, float64(vap.TxTotal),
				accessPointInfo.Name, vap.Name, vap.BSSID, vap.Radio, vap.RadioName, vap.ESSID, vap.Usage)

			// Station (client)
			for _, station := range vap.StationTable {
				ch <- prometheus.MustNewConstMetric(e.station.rxBytes, prometheus.CounterValue, float64(station.RxBytes),
					accessPointInfo.Name, vap.Name, station.Hostname, station.Mac)
				ch <- prometheus.MustNewConstMetric(e.station.txBytes, prometheus.CounterValue, float64(station.TxBytes),
					accessPointInfo.Name, vap.Name, station.Hostname, station.Mac)
				ch <- prometheus.MustNewConstMetric(e.station.noise, prometheus.CounterValue, float64(station.Noise),
					accessPointInfo.Name, vap.Name, station.Hostname, station.Mac)
				ch <- prometheus.MustNewConstMetric(e.station.signal, prometheus.CounterValue, float64(station.Signal),
					accessPointInfo.Name, vap.Name, station.Hostname, station.Mac)
			}
		}

	}

}
