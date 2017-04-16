# go-metrics Riemann Reporter

Library to report metrics collected with [go-metrics](https://github.com/rcrowley/go-metrics) to [Riemann](http://riemann.io).

## Example

Inside your application you can use the `riemann.Report` function to repeatedly report metrics. This will maintain a connection to Riemann (reconnecting in the event of a connection error) and report with the specified duration.

For example:

```go
package main

import (
	"github.com/pingles/go-metrics-riemann"
	"github.com/rcrowley/go-metrics"
	"log"
	"net"
	"time"
)

func main() {
    counter := metrics.NewCounter()
    metrics.Register("some counter", counter)
    
    go riemann.Report(metrics.DefaultRegistry, time.Second, "localhost:5555")
}
``` 

Alternatively, if you want to control the Riemann connection and when metrics are reported you can use `riemann.ReportOnce` and `riemann.RiemannConnect` in something like this:

```go
package main

import (
	"github.com/pingles/go-metrics-riemann"
	"github.com/rcrowley/go-metrics"
	"log"
	"net"
	"time"
)

func main() {
    counter := metrics.NewCounter()
    metrics.Register("some counter", counter)
    
    // RiemannConnect will block until it successfully connects
    client := riemann.RiemannConnect("localhost:5555")

    // ReportOnce will send all metrics once to a connected client. It
    // will return an error if there is an error whilst it sends the
    // metrics to Riemann.
    err := riemann.ReportOnce(metrics.DefaultRegistry, client)
    if err != nil {
        fmt.Println("Error reporting metrics to Riemann, connection error?")
    }
}
```

## License
Released under a BSD license. Please see LICENSE for more details.
