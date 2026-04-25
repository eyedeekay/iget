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
			parts := strings.SplitN(x, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				args.Header.Add(parts[0], parts[1])
			}
		}
	}
}
