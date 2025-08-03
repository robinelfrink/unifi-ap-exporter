package main

import (
	"flag"
	"os"

	unifiApExporter "unifi-ap-exporter/internal"

	log "github.com/sirupsen/logrus"
)

var (
	Version = "development"

	configPath = flag.String("config", "unifi-ap-exporter.yaml",
		"Configuration file")
	versionFlag = flag.Bool("version", false,
		"Show unifi-ap-exporter version")
	jsonLogging = flag.Bool("json", false,
		"Enable JSON logging")
	verboseLogging = flag.Bool("verbose", false,
		"Enable verbose logging")
	debugLogging = flag.Bool("debug", false,
		"Enable debug logging")
)

func main() {
	flag.Parse()

	if *jsonLogging {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}

	if *verboseLogging {
		log.SetLevel(log.DebugLevel)
	}

	if *debugLogging {
		log.SetLevel(log.TraceLevel)
	}

	if *versionFlag {
		log.Info("unifi-ap-exporter ", Version)
		os.Exit(0)
	}

	config, err := unifiApExporter.NewConfig(*configPath)
	if err != nil {
		log.Errorf("cannot read config file: %s", err)
		os.Exit(0)
	}

	collector := unifiApExporter.NewCollector(*config)
	exporter := unifiApExporter.NewExporter(*collector)
	exporter.Run()
}
