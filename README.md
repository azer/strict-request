# strict-request

Send HTTP requests with strict options, such as;

* Timeout
* Maximum content size (first given MBs)
* No redirects (Can be enabled back with `AllowRedirects` option)

## Example

```go
package main

import (
  "github.com/kozmos/strict-request"
)


func main () {
  resp, err := strictrequest.Get(testServer.URL, strictrequest.Options{
    TimeoutMs: 1000, // 1000 milliseconds
    MaxSizeMb: 0.5, // 500kb
  })

  if err != nil {
    panic(err)
  }
}
```

Check out tests for more details and documentation.
