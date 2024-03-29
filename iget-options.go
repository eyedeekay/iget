package iget

// Option is an IGet option
type Option func(*IGet)

// Lifespan is the lifespan to keep the destination alive
func Lifespan(i int) Option {
	return func(args *IGet) {
		args.destLifespan = i
	}
}

// MarkSize
func MarkSize(i int) Option {
	return func(args *IGet) {
		args.markSize = i
	}
}

// Linelength is the maximum length of an output line.
func LineLength(i int) Option {
	return func(args *IGet) {
		args.lineLength = i
	}
}

// Timeout defines the maximum timeout time
func Timeout(i int) Option {
	return func(args *IGet) {
		args.timeoutTime = i * 60 * 1000
	}
}

// Length is both the in and outbound tunnel length
func Length(i int) Option {
	return func(args *IGet) {
		args.tunnelLength = i
	}
}

// Inbound is the number of inbound tunnels to use
func Inbound(i int) Option {
	return func(args *IGet) {
		args.inboundTunnels = i
	}
}

// Outbound is the number of outbound tunnels to use
func Outbound(i int) Option {
	return func(args *IGet) {
		args.outboundTunnels = i
	}
}

// InboundBackups is the number of inbound backup tunnels to use
func InboundBackups(i int) Option {
	return func(args *IGet) {
		args.inboundBackups = i
	}
}

// OutboundBackups is the number of outbound backup tunnels to use
func OutboundBackups(i int) Option {
	return func(args *IGet) {
		args.outboundBackups = i
	}
}

// SamHost sets the host to look for a sam bridge at
func SamHost(i string) Option {
	return func(args *IGet) {
		args.samHost = i
	}
}

// SamPort sets the host to look for a sam port at
func SamPort(i string) Option {
	return func(args *IGet) {
		args.samPort = i
	}
}

// Verbose makes iget give verbose options
func Verbose(i bool) Option {
	return func(args *IGet) {
		args.verb = i
	}
}

// Debug sets the debug option in goSam
func Debug(i bool) Option {
	return func(args *IGet) {
		args.debug = i
	}
}

// Continue tells iget to resume a previously started download
func Continue(i bool) Option {
	return func(args *IGet) {
		args.continueDownload = i
	}
}

// URL sets the URL to retrieve
func URL(i string) Option {
	return func(args *IGet) {
		args.url = i
	}
}

// Username sets the SAM AUTH username
func Username(i string) Option {
	return func(args *IGet) {
		args.username = i
	}
}

// Password sets the SAM AUTH password
func Password(i string) Option {
	return func(args *IGet) {
		args.username = i
	}
}
