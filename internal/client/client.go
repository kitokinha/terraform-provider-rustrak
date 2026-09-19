package client

import (
	"net/http"
)

type RustrakClient struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

func NewRustrakClient(
	host string,
	token string,
) *RustrakClient {
	httpClient := &http.Client{}
	return &RustrakClient{
		HostURL:    host,
		Token:      token,
		HTTPClient: httpClient,
	}
}
