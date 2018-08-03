package main

import (
	"flag"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fiorix/go-smpp/smpp"
	"io"
	logger "log"
	"net/http"
	"os"
	"text/template"
)

var (
	log *logger.Logger
	tx  *smpp.Transmitter

	grafanaTmpl    *template.Template
	prometheusTmpl *template.Template

	configFile, logFile string
	debug               bool

	config struct {
		ListenAddr string `toml:"listen_address"`
		Sms        struct {
			Host     string `toml:"host"`
			Port     int    `toml:"port"`
			SystemID string `toml:"system_id"`
			Password string `toml:"password"`
		} `toml:"sms"`
		Grafana struct {
			Enable    bool     `toml:"enable"`
			Recipient []string `toml:"recipient"`
			From      string   `toml:"from"`
			User      string   `toml:"user"`
			Password  string   `toml:"password"`
			Template  string   `toml:"template"`
		} `toml:"grafana"`
		Prometheus struct {
			Enable    bool     `toml:"enable"`
			Recipient []string `toml:"recipient"`
			From      string   `toml:"from"`
			User      string   `toml:"user"`
			Password  string   `toml:"password"`
			Template  string   `toml:"template"`
		} `toml:"prometheus"`
		JSON struct {
			Enable   bool   `toml:"enable"`
			User     string `toml:"user"`
			Password string `toml:"password"`
		} `toml:"json"`
		Raw struct {
			Enable   bool   `toml:"enable"`
			User     string `toml:"user"`
			Password string `toml:"password"`
		} `toml:"raw"`
	}
)

func init() {
	flag.StringVar(&configFile, "c", "config.ini", "Config file")
	flag.StringVar(&logFile, "l", "", "Log file, if blank: not write log to file")
	flag.BoolVar(&debug, "d", false, "Debug on")
	flag.Parse()
}

func main() {
	var (
		err       error
		logWriter io.Writer
	)

	if logFile != "" {
		lf, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			logger.Printf("error opening log file: %v", err)
			os.Exit(2)
		}
		defer lf.Close()
		logWriter = io.MultiWriter(lf, os.Stdout)
	} else {
		logWriter = os.Stdout
	}
	log = logger.New(logWriter, "", logger.Ldate|logger.Ltime)

	if _, err = toml.DecodeFile(configFile, &config); err != nil {
		log.Printf("Parse config error: %s", err)
		os.Exit(2)
	}

	grafanaTmpl = template.Must(template.New("grafana").Parse(config.Grafana.Template))
	prometheusTmpl = template.Must(template.New("prometheus").Parse(config.Prometheus.Template))

	tx = &smpp.Transmitter{
		Addr:   fmt.Sprintf("%s:%d", config.Sms.Host, config.Sms.Port),
		User:   config.Sms.SystemID,
		Passwd: config.Sms.Password,
	}
	conn := tx.Bind()
	var status smpp.ConnStatus
	if status = <-conn; status.Error() != nil {
		log.Printf("Unable SMPP connect, aborting: %s", status.Error().Error())
	} else {
		log.Println("SMPP connection completed, status:", status.Status().String())
	}

	go func() {
		for c := range conn {
			if debug {
				log.Println("SMPP connection status:", c.Status())
			}
		}
	}()

	http.HandleFunc("/grafana", grafanaHandler)
	http.HandleFunc("/prometheus", prometheusHandler)
	http.HandleFunc("/json", jsonHandler)
	log.Printf("Listen on %s", config.ListenAddr)
	log.Printf("Listen error: %s", http.ListenAndServe(config.ListenAddr, nil))
}
