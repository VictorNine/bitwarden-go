package common

import (
	"io"
	"log"
	"net/http"
)

type Proxy struct {
	VaultURL string
}

func copyHeader(from, to http.Header) {
	for k, vv := range from {
		for _, v := range vv {
			to.Add(k, v)
		}
	}
}

// Proxy connections to vault server to avoid cros issues
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", p.VaultURL+r.URL.Path, nil)
	if err != nil {
		log.Fatal(err)
	}

	copyHeader(r.Header, req.Header)

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	copyHeader(resp.Header, w.Header())
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
