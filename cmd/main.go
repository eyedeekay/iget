package main

import (
	"flag"
	"fmt"
	i "github.com/eyedeekay/iget"
)

var (
	// application options
	samAddrString = flag.String("bridge-addr", "127.0.0.1", "host: of the SAM bridge")
	samPortString = flag.String("bridge-port", "7656", ":port of the SAM bridge")
	address       = flag.String("url", "", "i2p URL you want to retrieve")

	// debug options
	debugConnection = flag.Bool("conn-debug", false, "Print connection debug info")
	verboseLogging  = flag.Bool("verbose", false, "Print additional info about the process")
	output          = flag.String("output", "-", "Output path")

	//i2p options
	destLifespan    = flag.Int("lifespan", 6, "Lifespan of an idle i2p destination in minutes")
	timeoutTime     = flag.Int("timeout", 6, "Timeout duration in minutes")
	tunnelLength    = flag.Int("tunlength", 3, "Tunnel Length")
	inboundTunnels  = flag.Int("in-tunnels", 8, "Inbound Tunnel Count")
	outboundTunnels = flag.Int("out-tunnels", 8, "Inbound Tunnel Count")
	inboundBackups  = flag.Int("in-backups", 3, "Inbound Backup Count")
	outboundBackups = flag.Int("out-backups", 3, "Inbound Backup Count")

	// transport options
	keepAlives = flag.Bool("disable-keepalives", false, "Disable keepalives")
	idleConns  = flag.Int("idle-conns", 4, "Maximium idle connections per host")

	// request options
	method = flag.String("method", "GET", "Request method")
	closer = flag.Bool("close", true, "Close the request immediately after reading the response")
)

func main() {
	flag.Parse()
	if iget, ierr := i.NewIGet(
		i.Lifespan(*destLifespan),
		i.Timeout(*timeoutTime),
		i.Length(*tunnelLength),
		i.Inbound(*inboundTunnels),
		i.Outbound(*outboundTunnels),
		i.KeepAlives(*keepAlives),
		i.Idles(*idleConns),
		i.InboundBackups(*inboundBackups),
		i.OutboundBackups(*outboundBackups),
		i.Debug(*debugConnection),
		i.URL(*address),
		i.SamHost(*samAddrString),
		i.SamPort(*samPortString),
		i.Method(*method),
	); ierr != nil {
		fmt.Printf(ierr.Error())
	} else {
		if r, e := iget.Request(
			i.Close(*closer),
		); e != nil {
			fmt.Printf(ierr.Error())
		} else {
			if b, e := iget.Do(r); e != nil {
				fmt.Printf(e.Error())
			} else {
				iget.PrintResponse(b)
			}
		}
	}
}
