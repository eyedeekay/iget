package iget

// Option is an IGet option
type Option func(*IGet)

// Lifespan is the lifespan to keep the destination alive
func Lifespan(i int) Option {
	return func(args *IGet) {
		args.destLifespan = i
	}
}

// Timeout is the lifespan to keep the destination alive
func Timeout(i int) Option {
	return func(args *IGet) {
		args.timeoutTime = i
	}
}

// Length is the lifespan to keep the destination alive
func Length(i int) Option {
	return func(args *IGet) {
		args.tunnelLength = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func Inbound(i int) Option {
	return func(args *IGet) {
		args.inboundTunnels = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func Outbound(i int) Option {
	return func(args *IGet) {
		args.outboundTunnels = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func Idles(i int) Option {
	return func(args *IGet) {
		args.idleConns = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func InboundBackups(i int) Option {
	return func(args *IGet) {
		args.inboundBackups = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func OutboundBackups(i int) Option {
	return func(args *IGet) {
		args.outboundBackups = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func SamHost(i string) Option {
	return func(args *IGet) {
		args.samHost = i
	}
}

// Lifespan is the lifespan to keep the destination alive
func SamPort(i string) Option {
	return func(args *IGet) {
		args.samPort = i
	}
}

func Verbose(i bool) Option {
	return func(args *IGet) {
		args.verb = i
	}
}

func Debug(i bool) Option {
	return func(args *IGet) {
		args.debug = i
	}
}
