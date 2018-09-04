package iget

// Output sets the output to redirect to a file
func Output(i string) Option {
	return func(args *IGet) {
		args.outputPath = i
	}
}
