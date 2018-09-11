package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
)
import (
	i "github.com/eyedeekay/iget"
)

type arrayFlags []string

func (i *arrayFlags) String() string {
	var r string
	for _, s := range *i {
		r += s + ","
	}
	return strings.TrimSuffix(r, ",")
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	// application options
	samHostString  = flag.String("bridge-host", "127.0.0.1", "host: of the SAM bridge")
	samPortString  = flag.String("bridge-port", "7656", ":port of the SAM bridge")
	dsamAddrString = flag.String("bridge-addr", "127.0.0.1:7656", "host:port of the SAM bridge. Overrides bridge-host/bridge-port.")
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

	// compatibility options
	linelength  = flag.String("l", "", "Linelength(not enabled, provided so it doesn't break places where eepGet is already used, pipe it to something else to control line length, a wrapper will do this for iget)")
	linelength2 = flag.String("lineLength", "", "Linelength(not enabled, provided so it doesn't break places where eepGet is already used, pipe it to something else to control line length, a wrapper will do this for iget)")
	linelength3 = flag.String("-lineLength", "", "Linelength(not enabled, provided so it doesn't break places where eepGet is already used, pipe it to something else to control line length, a wrapper will do this for iget)")
	etag        = flag.String("e", "", "Set the etag header, not enabled yet, will break when used.")
	marksize    = flag.String("m", "", "Marksize(not enabled, provided so it doesn't break places where eepGet is already used)")
	retries     = flag.String("n", "", "Retries(not enabled yet, provided so it doesn't break places where eepGet is already used)")
	user        = flag.String("u", "", "Username for authenticating to SAM(not enabled yet, provided so it doesn't break places where eepGet is already used, will break non-empty usernames)")
	pass        = flag.String("x", "", "Password for authenticating to SAM(not enabled yet, provided so it doesn't break places where eepGet is already used, will break non-empty passwords)")
)

var (
	samAddrString = "127.0.0.1:7656"
	timeoutTime   = 6
	output        = "-"
)

func main() {
	var headers arrayFlags
	stimeoutTime := flag.Int("t", 6, "Timeout duration in minutes")
	soutput := flag.String("o", "-", "Output path")
	ssamAddrString := flag.String("p", "127.0.0.1:7656", "host:port of the SAM bridge. Overrides bridge-host/bridge-port.")
	flag.Var(&headers, "h", "Add a header to the request in the form key=value")
	flag.Parse()
	if args := flag.Args(); len(args) == 1 {
		*address = args[0]
	}
	if *address == "" {
		log.Fatal("Fatal error, no url supplied by user (pass -url or an address argument)")
	}
	if _, err := url.ParseRequestURI(*address); err != nil {
		temp := "http://" + *address
		*address = temp
	}
	if *soutput != "-" {
		output = *soutput
	} else {
		output = *doutput
	}
	if *stimeoutTime != 6 {
		timeoutTime = *stimeoutTime
	} else {
		timeoutTime = *dtimeoutTime
	}
	if *ssamAddrString != "127.0.0.1:7656" {
		samAddrString = *ssamAddrString
	} else {
		samAddrString = *dsamAddrString
	}
	if strings.Contains(samAddrString, "4444") {
		fmt.Printf("%s", "This application uses the SAM API instead of the http proxy.")
		fmt.Printf("%s", "Please modify your scripts to use the SAM port.")
		return
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
