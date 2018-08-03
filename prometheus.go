package main

import (
	"bytes"
	"encoding/json"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"io/ioutil"
	"net/http"
)

type prometheusType struct {
	Version     string `json:"version"`
	GroupKey    string `json:"groupKey"`
	Status      string `json:"status"`
	Receiver    string `json:"receiver"`
	GroupLabels map[string]string `json:"groupLabels"`
	CommonLabels map[string]string `json:"commonLabels"`
	CommonAnnotations map[string]string `json:"commonAnnotations"`
	ExternalURL string `json:"externalURL"`
	Alerts      []struct {
		Status string `json:"status"`
		Labels map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
		StartsAt     string `json:"startsAt"`
		EndsAt       string `json:"endsAt"`
		GeneratorURL string `json:"generatorURL"`
	} `json:"alerts"`
}

func prometheusHandler(w http.ResponseWriter, r *http.Request) {
	if !config.Prometheus.Enable {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, password, ok := r.BasicAuth()
	if !(ok && user == config.Prometheus.User && password == config.Prometheus.Password) {
		log.Printf("Bad auth to prometheus from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var prometheus prometheusType

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &prometheus)
	if err != nil {
		log.Printf("Unmarshal prometheus json error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sms := &bytes.Buffer{}
	err = prometheusTmpl.Execute(sms, prometheus)
	if err != nil {
		log.Printf("Parse prometheus template error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range config.Prometheus.Recipient {
		_, err = tx.SubmitLongMsg(&smpp.ShortMessage{
			Src:      config.Prometheus.From,
			Dst:      config.Prometheus.Recipient[i],
			Text:     pdutext.UCS2(sms.Bytes()),
			Register: pdufield.NoDeliveryReceipt,
		})
		if err == smpp.ErrNotConnected {
			log.Printf("Smpp not connected for send to %s", config.Prometheus.Recipient[i])
			http.Error(w, "Oops.", http.StatusServiceUnavailable)
			return
		}
		if err != nil {
			log.Printf("Sms send to %s error: %s", config.Prometheus.Recipient[i], err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
