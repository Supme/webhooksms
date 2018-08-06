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

type grafanaType struct {
	Title       string `json:"title"`
	RuleID      int    `json:"ruleId"`
	RuleName    string `json:"ruleName"`
	RuleURL     string `json:"ruleUrl"`
	State       string `json:"state"`
	ImageURL    string `json:"imageUrl"`
	Message     string `json:"message"`
	EvalMatches []struct {
		Metric string            `json:"metric"`
		Tags   map[string]string `json:"tags"`
		Value  float64           `json:"value"`
	} `json:"evalMatches"`
}

func grafanaHandler(w http.ResponseWriter, r *http.Request) {
	if !config.Grafana.Enable {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	user, password, ok := r.BasicAuth()
	if !(ok && user == config.Grafana.User && password == config.Grafana.Password) {
		log.Printf("Bad auth to grafana from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var grafana grafanaType

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &grafana)
	if err != nil {
		log.Printf("Unmarshal grafana json error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sms := &bytes.Buffer{}
	err = grafanaTmpl.Execute(sms, grafana)
	if err != nil {
		log.Printf("Parse grafana template error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	for i := range config.Grafana.Recipient {
		_, err = tx.SubmitLongMsg(&smpp.ShortMessage{
			Src:      config.Grafana.From,
			Dst:      config.Grafana.Recipient[i],
			Text:     pdutext.UCS2(sms.Bytes()),
			Register: pdufield.NoDeliveryReceipt,
		})
		if err == smpp.ErrNotConnected {
			log.Printf("Smpp not connected for send to %s", config.Grafana.Recipient[i])
			http.Error(w, "Oops.", http.StatusServiceUnavailable)
			return
		}
		if err != nil {
			log.Printf("Sms send to %s error: %s", config.Grafana.Recipient[i], err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
}
