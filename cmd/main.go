package main

import (
	"flag"
	"fmt"
	"strings"
)
import (
	i "github.com/eyedeekay/iget"
)

var (
	// application options
	samHostString  = flag.String("bridge-host", "127.0.0.1", "host: of the SAM bridge")
	samPortString  = flag.String("bridge-port", "7656", ":port of the SAM bridge")
	dsamAddrString = flag.String("bridge-addr", "127.0.0.1:7656", "host:port of the SAM bridge. Inactive at the moment.")
	address        = flag.String("url", "", "i2p URL you want to retrieve")

	// debug options
	debugConnection = flag.Bool("conn-debug", false, "Print connection debug info")
	verboseLogging  = flag.Bool("verbose", false, "Print additional info about the process")
	doutput         = flag.String("output", "-", "Output path")

	//i2p options
	destLifespan    = flag.Int("lifespan", 6, "Lifespan of an idle i2p destination in minutes")
	dtimeoutTime    = flag.Int("timeout", 6, "Timeout duration in minutes")
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

var (
	samAddrString = "127.0.0.1:7656"
	timeoutTime   = 6
	output        = "-"
)

func main() {
	stimeoutTime := flag.Int("t", 6, "Timeout duration in minutes")
	soutput := flag.String("o", "-", "Output path")
	ssamAddrString := flag.String("p", "127.0.0.1:7656", "host:port of the SAM bridge. Inactive at the moment.")
	flag.Parse()
	if *soutput != "-" {
		output = soutput
	} else {
		output = doutput
	}
	if *stimeoutTime != 6 {
		timeoutTime = stimeoutTime
	} else {
		timeoutTime = dtimeoutTime
	}
	if *ssamAddrString != "127.0.0.1:7656" {
		samAddrString = ssamAddrString
	} else {
		samAddrString = dsamAddrString
	}
	if samAddrString != *samHostString+":"+*samPortString && samAddrString != "127.0.0.1:7656" {
		x := strings.Split(samAddrString, ":")
		if len(x) == 2 {
			*samHostString = x[0]
			*samPortString = x[1]
		} else if len(x) == 1 {
			*samPortString = x[0]
		}
	}
	if iget, ierr := i.NewIGet(
		i.Lifespan(*destLifespan),
		i.Timeout(timeoutTime),
		i.Length(*tunnelLength),
		i.Inbound(*inboundTunnels),
		i.Outbound(*outboundTunnels),
		i.KeepAlives(*keepAlives),
		i.Idles(*idleConns),
		i.InboundBackups(*inboundBackups),
		i.OutboundBackups(*outboundBackups),
		i.Debug(*debugConnection),
		i.URL(*address),
		i.SamHost(*samHostString),
		i.SamPort(*samPortString),
		i.Method(*method),
		i.Output(output),
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
