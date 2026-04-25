# iget

An I2P HTTP client for use as a Go library and CLI tool. It communicates exclusively over I2P via the SAM bridge API — no HTTP proxy, no clearnet leakage.

## Install

```
go install github.com/eyedeekay/iget/iget@latest
```

## CLI Usage

```
iget [options] <url>

Options:
  -url string          URL to retrieve (or pass as a positional argument)
  -o string            Output path (default: stdout)
  -output string       Output path (long form)
  -method string       HTTP method (default: GET)
  -c                   Resume a partial download (default: true)
  -close               Close connection after response (default: true)
  -m int               Show download progress (any value > 0 enables)
  -l int               Line length of output (default: 80)
  -t int               Timeout in minutes (default: 6)
  -timeout int         Timeout in minutes (long form)
  -n int               Retry count (default: 3)
  -h key=value         Add request header (repeatable)
  -header key=value    Add request header (long form, repeatable)

SAM Bridge:
  -bridge-host string  SAM bridge host (default: 127.0.0.1)
  -bridge-port string  SAM bridge port (default: 7656)
  -bridge-addr string  SAM bridge host:port, overrides above (default: 127.0.0.1:7656)
  -p string            SAM bridge host:port (short form)

I2P Tunnel:
  -lifespan int        Idle destination lifespan in minutes (default: 6)
  -tunlength int       Tunnel length (default: 3)
  -in-tunnels int      Inbound tunnel count (default: 8)
  -out-tunnels int     Outbound tunnel count (default: 8)
  -in-backups int      Inbound backup tunnel count (default: 3)
  -out-backups int     Outbound backup tunnel count (default: 3)

Transport:
  -disable-keepalives  Disable HTTP keep-alives
  -idle-conns int      Max idle connections per host (default: 4)

Debug:
  -verbose             Print process information to stderr
  -conn-debug          Print SAM connection debug info
```

**Note:** iget uses the SAM bridge API (default port `7656`), not the I2P HTTP proxy (port `4444`). Passing a port `4444` address will be rejected with an error.

## Library Usage

```go
import iget "github.com/eyedeekay/iget"
```

### Construct a client

```go
client, err := iget.NewIGet(
    iget.SamHost("127.0.0.1"),
    iget.SamPort("7656"),
    iget.URL("http://example.i2p/"),
    iget.Timeout(6),
    iget.Inbound(4),
    iget.Outbound(4),
)
```

### Make a request

```go
req, err := client.Request()
resp, err := client.Do(req)
client.PrintResponse(resp)
```

### As an `http.RoundTripper`

`IGet` implements `http.RoundTripper`, so it can be used as a transport for any standard or third-party HTTP client:

```go
httpClient := &http.Client{Transport: client}
```

Or with libraries like [Resty](https://github.com/go-resty/resty):

```go
restyClient := resty.New().SetTransport(client)
```

### As a Heimdall-compatible client

`IGet.Do` satisfies [Heimdall's `Client` interface](https://github.com/gojek/heimdall) directly.

### Options reference

| Option | Description |
|---|---|
| `SamHost(string)` | SAM bridge host |
| `SamPort(string)` | SAM bridge port |
| `URL(string)` | Default request URL |
| `Timeout(int)` | Timeout in minutes |
| `Lifespan(int)` | Idle destination lifespan in minutes |
| `Length(int)` | Tunnel length (in and out) |
| `Inbound(int)` | Inbound tunnel count |
| `Outbound(int)` | Outbound tunnel count |
| `InboundBackups(int)` | Inbound backup tunnel count |
| `OutboundBackups(int)` | Outbound backup tunnel count |
| `KeepAlives(bool)` | Enable HTTP keep-alives |
| `Idles(int)` | Max idle connections per host |
| `Output(string)` | Output file path (`-` for stdout) |
| `LineLength(int)` | Output line wrap length |
| `MarkSize(int)` | Enable download progress display |
| `Continue(bool)` | Resume partial downloads |
| `Verbose(bool)` | Verbose logging |
| `Debug(bool)` | SAM connection debug logging |
| `Username(string)` | SAM AUTH username |
| `Password(string)` | SAM AUTH password |

## Build

```
make deps build
```