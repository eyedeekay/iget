package iget

// Option is an IGet option
type Option func(*IGet)

// Lifespan is the lifespan to keep the destination alive, in minutes.
func Lifespan(i int) Option {
	return func(args *IGet) {
		args.destLifespan = i * 60 * 1000
	}
}

// MarkSize sets the download progress interval in bytes. A progress line is
// printed to stdout after every MarkSize bytes are received. Set to 0 (the
// default) to disable progress output.
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

// Debug enables verbose debug logging
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

// Username sets the SAM AUTH username for connecting to a SAM bridge that
// requires SAM v3.2+ USER/PASSWORD authentication.
func Username(i string) Option {
	return func(args *IGet) {
		args.username = i
	}
}

// Password sets the SAM AUTH password for connecting to a SAM bridge that
// requires SAM v3.2+ USER/PASSWORD authentication.
func Password(i string) Option {
	return func(args *IGet) {
		args.password = i
	}
}

// Body sets the request body for non-GET methods (e.g., POST, PUT).
func Body(i string) Option {
	return func(args *IGet) {
		args.body = i
	}
}

// SessionName sets the SAM session name used to identify this client's I2P
// destination. By default NewIGet generates a unique per-invocation name so
// that consecutive runs use different destinations and cannot be correlated.
// Set an explicit name when a persistent, stable I2P identity is intentional.
func SessionName(name string) Option {
	return func(args *IGet) {
		args.sessionName = name
	}
}

// Output sets the output to redirect to a file
func Output(i string) Option {
	return func(args *IGet) {
		args.outputPath = i
	}
}

// ToPort sets the SAM virtual destination port for outgoing connections.
func ToPort(s string) Option {
	return func(args *IGet) {
		args.toPort = s
	}
}

// FromPort sets the SAM virtual source port for incoming connections.
func FromPort(s string) Option {
	return func(args *IGet) {
		args.fromPort = s
	}
}

// DisableKeepAlives disables keepalives on the HTTP transport when set to true
func DisableKeepAlives(i bool) Option {
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
