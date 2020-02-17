package main

import "os"

const defaultAddress = ":1337"

func main() {
	// grab a bind address from the 1st cmdline arg
	address := defaultAddress
	if len(os.Args) > 1 {
		address = os.Args[1]
	}

	startServer(address)
}
