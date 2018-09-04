package iget

import (
	"net/http"
)

// RequestOption is helpful for generating requests
type RequestOption func(*http.Request)

// Method sets the request method
func Method(i string) Option {
	return func(args *IGet) {
		args.method = i
	}
}

// Close sets the http.Request Close variable
func Close(i bool) RequestOption {
	return func(args *http.Request) {
		args.Close = i
	}
}

/*
// Header adds a request header
func Header(i string) RequestOption {
	return func(args *http.Request) {
		args.Header = i
	}
}
*/

/*
// OPT sets the request opt
func OPT(i string) RequestOption {
	return func(args *http.Request) {
		args.OPT = i
	}
}
*/
