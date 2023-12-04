package httphandling

import (
	"log"
	"net/http"
)

func CreateReqFromReq(req *http.Request, newURL string) (*http.Request, http.Client) {
	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	freq, reqerror := http.NewRequest(req.Method, newURL, req.Body)
	if reqerror != nil {
		log.Fatalln(reqerror)
	}

	freq.Header = req.Header

	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	//
	// https://go.dev/src/net/http/transport.go#L190
	freq.Header.Del("Accept-Encoding")

	return freq, client
}
