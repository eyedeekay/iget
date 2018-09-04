package main

import (
	"flag"
	"fmt"
	i "github.com/eyedeekay/iget"
)

var (
	samAddrString = flag.String("bridge-addr", "127.0.0.1", "host: of the SAM bridge")
	samPortString = flag.String("bridge-port", "7656", ":port of the SAM bridge")
	address       = flag.String("url", "", "i2p URL you want to retrieve")

	debugConnection = flag.Bool("conn-debug", false, "Print connection debug info")
	verboseLogging  = flag.Bool("verbose", false, "Print connection debug info")

	destLifespan    = flag.Int("lifespan", 6, "Lifespan of an idle i2p destination in minutes(default six)")
	timeoutTime     = flag.Int("timeout", 6, "Timeout duration in minutes(default six)")
	tunnelLength    = flag.Int("tunlength", 3, "Tunnel Length(default 3)")
	inboundTunnels  = flag.Int("in-tunnels", 8, "Inbound Tunnel Count(default 8)")
	outboundTunnels = flag.Int("out-tunnels", 8, "Inbound Tunnel Count(default 8)")
	keepAlives      = flag.Bool("disable-keepalives", false, "Disable keepalives(default false)")
	idleConns       = flag.Int("idle-conns", 4, "Maximium idle connections per host(default 4)")
	inboundBackups  = flag.Int("in-backups", 3, "Inbound Backup Count(default 3)")
	outboundBackups = flag.Int("out-backups", 3, "Inbound Backup Count(default 3)")
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
	); ierr != nil {
		fmt.Printf(ierr.Error())
	} else {
		if r, e := iget.Request(); e != nil {
			fmt.Printf(ierr.Error())
		} else {
			if b, e := iget.DoString(r); e != nil {
				fmt.Printf(e.Error())
			} else {
				fmt.Printf("%s", b)
			}
		}
	}
}
