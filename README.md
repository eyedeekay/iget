# iget
i2p terminal http client.

## description:
This is a highly-configurable curl/wget like client which exclusively works
over i2p. It works via the SAM API which means it has some advantages and
some disadvantages, as follows:

Wherever possible, short arguments will mirror their curl equivalents.
However, I'm not trying to implement every single curl option, and if
there are arguments that are labeled differently between curl and eepget,
eepget options will be used instead. I haven't decided if I want it to be
able to spider eepsites on it's own, but I'm leaning toward no. That's what
lynx and grep are for.

### Advantages:
These advantages motivated development. More may emerge as it continues.

  - uses the SAM API to prevent destination-reuse for different sites
  - uses the SAM API directly(not forwarding) so it can't leak information
    to clearnet services
  - inline options to configure i2cp, so for example we can have 8 tunnels
    in and 2 tunnels out

### Disadvantages:
Only two I know of so far.

  - marginally slower, due to tunnel-creation at runtime.
  - a few missing options compared to eepget

## to build:

        make deps build

## to use:

```
Usage of ./bin/iget:
  -bridge-addr string
    	host:port of the SAM bridge. Overrides bridge-host/bridge-port. (default "127.0.0.1:7656")
  -bridge-host string
    	host: of the SAM bridge (default "127.0.0.1")
  -bridge-port string
    	:port of the SAM bridge (default "7656")
  -close
    	Close the request immediately after reading the response (default true)
  -conn-debug
    	Print connection debug info
  -disable-keepalives
    	Disable keepalives
  -idle-conns int
    	Maximium idle connections per host (default 4)
  -in-backups int
    	Inbound Backup Count (default 3)
  -in-tunnels int
    	Inbound Tunnel Count (default 8)
  -lifespan int
    	Lifespan of an idle i2p destination in minutes (default 6)
  -method string
    	Request method (default "GET")
  -o string
    	Output path (default "-")
  -out-backups int
    	Inbound Backup Count (default 3)
  -out-tunnels int
    	Inbound Tunnel Count (default 8)
  -output string
    	Output path (default "-")
  -p string
    	host:port of the SAM bridge. Overrides bridge-host/bridge-port. (default "127.0.0.1:7656")
  -t int
    	Timeout duration in minutes (default 6)
  -timeout int
    	Timeout duration in minutes (default 6)
  -tunlength int
    	Tunnel Length (default 3)
  -url string
    	i2p URL you want to retrieve
  -verbose
    	Print additional info about the process
```

