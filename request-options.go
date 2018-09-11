package iget

import (
	"net/http"
	"strings"
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

// Headers adds a request header
func Headers(i []string) RequestOption {
	return func(args *http.Request) {
		for _, x := range i {
			//args.headers = append(args.headers, x)
			if len(strings.Split(x, "=")) == 2 {
				args.Header.Add(strings.Split(x, "=")[0], strings.Split(x, "=")[1])
			}
		}
	}
}

/*
// OPT sets the request opt
func OPT(i string) RequestOption {
	return func(args *http.Request) {
		args.OPT = i
	}
}
*/
