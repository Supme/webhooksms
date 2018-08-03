package main

import (
	"encoding/json"
	"github.com/fiorix/go-smpp/smpp"
	"github.com/fiorix/go-smpp/smpp/pdu/pdufield"
	"github.com/fiorix/go-smpp/smpp/pdu/pdutext"
	"io/ioutil"
	"net/http"
)

type jsonType struct {
	Cmd   string   `json:"cmd"`
	User  string   `json:"user"`
	Pass  string   `json:"pass"`
	From  string   `json:"from"`
	Recip []string `json:"recip"`
	Text  string   `json:"text"`
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	if !config.JSON.Enable {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	var js jsonType

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &js)
	if err != nil {
		log.Printf("Unmarshal json json error: %s", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if user, pass, ok := r.BasicAuth(); ok {
		js.User = user
		js.Pass = pass
	}

	if js.User != config.JSON.User && js.Pass != config.JSON.Password {
		log.Printf("Bad auth to json from %s", r.RemoteAddr)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	switch js.Cmd {
	case "send":
		for i := range js.Recip {
			_, err = tx.SubmitLongMsg(&smpp.ShortMessage{
				Src:      js.From,
				Dst:      js.Recip[i],
				Text:     pdutext.UCS2(js.Text),
				Register: pdufield.NoDeliveryReceipt,
			})
			if err == smpp.ErrNotConnected {
				log.Printf("Smpp not connected for send to %s", js.Recip[i])
				http.Error(w, "Oops.", http.StatusServiceUnavailable)
				return
			}
			if err != nil {
				log.Printf("Sms send to %s error: %s", js.Recip[i], err.Error())
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}
	default:
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}
