# Net Limit
Net Limit is a programming toolkit for limiting the network transfer rate. It wrap net.Conn with rate limiter to achieve limited transfer rate.

## Background
Given File URL to download, I wanted to limit the transfer rate so that the remaining bandwith can be used for someting else. Limiting the reading of resp.Body will only limit the reading rate of the resp.Body but the underlying connection is still buffering the data from the server. Hence, creating limiter in connection level is neccesary.

## Usages
### Limit net.Conn Transfer Rate

```go
package main

import (
    "net"

    "github.com/muktihari/netlimit"
    "golang.org/x/time/rate"
)

func main() {
    conn, err := net.Dial("tcp4", ":8080")
    if err != nil {
        panic(err)
    }

    limit := 100 << 10 // 100 KB /s 
    ratelimit := rate.NewLimiter(rate.Limit(limit), limit)
    connlimit := netlimit.NewConn(conn, ratelimit, nil) // limit read from the server

    ...
}
```

### Limit HTTP Transfer Rate
#### Example: File Downloader with Limit
```go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	httplimit "github.com/muktihari/netlimit/http"
	"golang.org/x/time/rate"
)

func main() {
    var (
        limit      = 100 << 10 // 100 KB/s
        ratelimit  = rate.NewLimiter(rate.Limit(limit), limit)
        transport  = httplimit.NewTransport(nil, ratelimit, nil)
        httpclient = &http.Client{Transport: transport}
    )
	
	req, err := http.NewRequest(http.MethodGet, "<< url >>", nil)
	if err != nil {
		panic(err)
	}

	resp, err := httpclient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("exited: %s", resp.Status)
		os.Exit(1)
	}

	f, err := os.OpenFile("downloaded-file", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0777)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	if err != nil {
		panic(err)
	}
}

```