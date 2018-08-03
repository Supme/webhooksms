package main

import "net/http"

// ToDo
type rawType struct {}

func rawHandler(w http.ResponseWriter, r *http.Request) {
	if !config.Raw.Enable {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}
}