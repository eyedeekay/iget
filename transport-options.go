package iget

// KeepAlives tells to allow keepalives
func KeepAlives(i bool) Option {
	return func(args *IGet) {
		args.keepAlives = i
	}
}

// Idles tells the max number of idle connections to allow
func Idles(i int) Option {
	return func(args *IGet) {
		args.idleConns = i
	}
}
