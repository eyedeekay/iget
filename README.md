# iget
i2p terminal http client.

## description:
This is a highly-configurable curl/wget like client which exclusively works
over i2p. It works via the SAM API which means it has some advantages and
some disadvantages, as follows:

### Advantages:
These advantages motivated development. More may emerge as it continues.

  - uses the SAM API to prevent destination-reuse for different sites
  - uses the SAM API directly(not forwarding) so it can't leak information
    to clearnet services
  - inline options to configure i2cp, so for example we can have 8 tunnels
    in and 2 tunnels out

### Disadvantages:
Only one I know of so far.

  - marginally slower, due to tunnel-creation at runtime.

### Security Notes:

  - TLS certificate verification is disabled (`InsecureSkipVerify: true`). This
    is intentional: I2P eepsites commonly use self-signed certificates and
    certificate authorities have no jurisdiction over the I2P network. All
    traffic is already routed through the I2P network via SAM, so there is no
    clearnet exposure. Users who require strict certificate pinning should be
    aware of this behaviour.

Wherever possible, short arguments will mirror their curl equivalents.
However, I'm not trying to implement every single curl option, and if
there are arguments that are labeled differently between curl and eepget,
eepget options will be used instead.

## to build:

        make deps build

## to use:

```
iget is a highly-configurable curl/wget-like client that works exclusively over i2p via the SAM API.

Usage:
  iget [URL] [flags]

Flags:
  -p, --bridge-addr string   host:port of the SAM bridge (overrides bridge-host/bridge-port)
      --bridge-host string   host of the SAM bridge (default "127.0.0.1")
      --bridge-port string   port of the SAM bridge (default "7656")
      --close                close the request immediately after reading the response (default true)
      --config string        config file (default: $HOME/.iget.yaml)
      --conn-debug           print connection debug info
  -c, --continue             resume file from previous download (default true)
  -d, --data string          request body for POST/PUT
      --disable-keepalives   disable keepalives
  -e, --etag string          set the If-None-Match request header for conditional GETs
      --from-port string     SAM virtual source port
  -H, --header stringArray   add a request header in key=value form (repeatable)
  -h, --help                 help for iget
      --idle-conns int       maximum idle connections per host (default 4)
      --in-backups int       inbound backup count (default 3)
      --in-tunnels int       inbound tunnel count (default 8)
      --lifespan int         lifespan of an idle i2p destination in minutes (default 6)
  -l, --line-length int      control line length of output (0 = unlimited) (default 80)
  -m, --mark-size int        show download progress (any value > 0 enables)
      --method string        request method (default "GET")
      --out-backups int      outbound backup count (default 3)
      --out-tunnels int      outbound tunnel count (default 8)
  -o, --output string        output path (- for stdout) (default "-")
  -x, --password string      password for SAM authentication
  -n, --retries int          number of retries (default 3)
  -t, --timeout int          timeout duration in minutes (default 6)
      --to-port string       SAM virtual destination port
      --tunlength int        tunnel length (default 3)
      --url string           i2p URL to retrieve
  -u, --username string      username for SAM authentication
  -v, --verbose              print additional info about the process
```

