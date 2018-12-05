package main

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func basicAuth(username, password string, pass http.HandlerFunc) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed no auth", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !validate(username, password, pair[0], pair[1]) {
			http.Error(w, "authorization failed invalid pass", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

func validate(refu, refp, u, p string) bool {
	if u == refu && p == refp {
		return true
	}
	return false
}
