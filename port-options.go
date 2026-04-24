package iget

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
