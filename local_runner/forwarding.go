package localrunner

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func forward_request(prefix string, baseUrl string, w http.ResponseWriter, r *http.Request) {
	origHost := r.URL.Host

	u, err := url.Parse(baseUrl + strings.TrimPrefix(r.RequestURI, prefix))
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	r.URL = u
	r.RequestURI = ""

	resp, err := http.DefaultClient.Do(r)
	if resp == nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}

	w.Header().Set("X-Forwarded-For", r.RemoteAddr)
	w.Header().Set("X-Forwarded-Proto", r.Proto)
	w.Header().Set("X-Forwarded-Host", origHost)

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
