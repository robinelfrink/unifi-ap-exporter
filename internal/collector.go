package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type Collector struct {
	config Config
}

type APSystemStats struct {
	CPU float64 `json:"cpu,string"`
	Mem float64 `json:"mem,string"`
}

type APSysStats struct {
	LoadAvg1  float64 `json:"loadavg_1,string"`
	LoadAvg5  float64 `json:"loadavg_5,string"`
	LoadAvg15 float64 `json:"loadavg_15,string"`
	MemUsed   int64   `json:"mem_used"`
	MemTotal  int64   `json:"mem_total"`
	MemBuffer int64   `json:"mem_buffer"`
}

type APInterface struct {
	TxBytes int64 `json:"tx_bytes"`
	RxBytes int64 `json:"rx_bytes"`
	Up      bool  `json:"up"`
}

type APRadio struct {
	Radio              string   `json:"radio"`
	RadioName          string   `json:"name"`
	CurrentAntennaGain int64    `json:"builtin_ant_gain"`
	MaxTxpower         int64    `json:"max_txpower"`
	MinTxpower         int64    `json:"min_txpower"`
	ScanTable          []APScan `json:"scan_table"`
}

type APVap struct {
	Name         string      `json:"name"`
	Radio        string      `json:"radio"`
	RadioName    string      `json:"radio_name"`
	RxBytes      int64       `json:"rx_bytes"`
	RxDropped    int64       `json:"rx_dropped"`
	RxErrors     int64       `json:"rx_errors"`
	TxBytes      int64       `json:"tx_bytes"`
	TxDropped    int64       `json:"tx_dropped"`
	TxErrors     int64       `json:"tx_errors"`
	TxPower      int64       `json:"tx_power"`
	TxRetries    int64       `json:"tx_retries"`
	TxSuccess    int64       `json:"tx_success"`
	TxTotal      int64       `json:"tx_total"`
	Channel      int64       `json:"channel"`
	BSSID        string      `json:"bssid"`
	ESSID        string      `json:"essid"`
	Usage        string      `json:"usage"`
	StationTable []APStation `json:"sta_table"`
}

type APStation struct {
	Hostname string `json:"hostname"`
	Mac      string `json:"mac"`
	TxBytes  int64  `json:"tx_bytes"`
	RxBytes  int64  `json:"rx_bytes"`
	Noise    int64  `json:"noise"`
	Signal   int64  `json:"signal"`
}

type APScan struct {
	BSSID     string `json:"bssid"`
	Channel   int64  `json:"channel"`
	ESSID     string `json:"essid"`
	Frequency int64  `json:"freq"`
	Noise     int64  `json:"noise"`
	Security  string `json:"security"`
	Signal    int64  `json:"signal"`
}

type AccessPointInfo struct {
	IP             string `json:"ip"`
	Mac            string `json:"mac"`
	Model          string `json:"model"`
	ModelName      string `json:"model_display"`
	Name           string
	Serial         string        `json:"serial"`
	Version        string        `json:"version"`
	Uptime         int64         `json:"uptime"`
	SystemStats    APSystemStats `json:"system-stats"`
	SysStats       APSysStats    `json:"sys_stats"`
	InterfaceTable []APInterface `json:"if_table"`
	RadioTable     []APRadio     `json:"radio_table"`
	VAPTable       []APVap       `json:"vap_table"`
	Value          float64
}

func NewCollector(config Config) *Collector {
	collector := &Collector{
		config: config,
	}
	return collector
}

func (c *Collector) Collect() (*[]AccessPointInfo, error) {
	var accessPointInfos = []AccessPointInfo{}

	for _, accessPoint := range c.config.AccessPoints {
		accessPointInfo, err := c.Fetch(accessPoint)
		if err != nil {
			accessPointInfos = append(accessPointInfos, AccessPointInfo{
				Name:  accessPoint.Name,
				IP:    accessPoint.Address,
				Value: 0,
			})
			log.Errorf("%s: %s", accessPoint.Address, err)
			continue
		}

		accessPointInfos = append(accessPointInfos, *accessPointInfo)
	}

	return &accessPointInfos, nil
}

func (c *Collector) Fetch(accessPoint AccessPointConfig) (*AccessPointInfo, error) {
	config := &ssh.ClientConfig{
		User:            accessPoint.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// Use password as authentication option
	if accessPoint.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(accessPoint.Password))
	}
	// Use key as authentication option
	if accessPoint.KeyFile != "" {
		key, err := os.ReadFile(accessPoint.KeyFile)
		if err != nil {
			return nil, err
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(signer))
	}

	client, err := ssh.Dial("tcp", fmt.Sprint(accessPoint.Address, ":22"), config)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = client.Close()
	}()

	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = session.Close()
	}()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// "mca" stands for "Management Control Agent"
	if err := session.Run("mca-dump"); err != nil {
		return nil, err
	}
	output, err := io.ReadAll(stdout)
	if err != nil {
		return nil, err
	}

	accessPointInfo := &AccessPointInfo{
		Name:  accessPoint.Name,
		Value: 1,
	}
	if err = json.Unmarshal(output, &accessPointInfo); err != nil {
		return nil, err
	}

	return accessPointInfo, err
}
