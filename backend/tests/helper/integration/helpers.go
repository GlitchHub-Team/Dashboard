package integration

import (
	"net/http"
)

func AuthHeader(jwt string) http.Header {
	header := http.Header{}
	header.Set("Authorization", "Bearer "+jwt)
	return header
}