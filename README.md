# iget
i2p terminal http client.

## to build:

        make deps build

## to use:

```
Usage of ./bin/iget:
  -bridge-addr string
    	host: of the SAM bridge (default "127.0.0.1")
  -bridge-port string
    	:port of the SAM bridge (default "7656")
  -close
    	Close the request immediately after reading the response (default true)
  -conn-debug
    	Print connection debug info
  -disable-keepalives
    	Disable keepalives(default false)
  -idle-conns int
    	Maximium idle connections per host(default 4) (default 4)
  -in-backups int
    	Inbound Backup Count(default 3) (default 3)
  -in-tunnels int
    	Inbound Tunnel Count(default 8) (default 8)
  -lifespan int
    	Lifespan of an idle i2p destination in minutes(default six) (default 6)
  -method string
    	Request method (default "GET")
  -out-backups int
    	Inbound Backup Count(default 3) (default 3)
  -out-tunnels int
    	Inbound Tunnel Count(default 8) (default 8)
  -timeout int
    	Timeout duration in minutes(default six) (default 6)
  -tunlength int
    	Tunnel Length(default 3) (default 3)
  -url string
    	i2p URL you want to retrieve
  -verbose
    	Print connection debug info
```

