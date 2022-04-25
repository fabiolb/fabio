# circonusllhist

A golang implementation of the OpenHistogram [libcircllhist](https://github.com/openhistogram/libcircllhist) library.

[![Go Reference](https://pkg.go.dev/badge/github.com/openhistogram/circonusllhist.svg)](https://pkg.go.dev/github.com/openhistogram/circonusllhist)

## Overview

Package `circllhist` provides an implementation of OpenHistogram's fixed log-linear histogram data structure.  This allows tracking of histograms in a composable way such that accurate error can be reasoned about.

## License

[Apache 2.0](LICENSE)

## Documentation

More complete docs can be found at [pkg.go.dev](https://pkg.go.dev/github.com/openhistogram/circonusllhist)

## Usage Example

```
package main

import (
    "fmt"
    "github.com/openhistogram/circonusllhist"
)

func main() {
    //Create a new histogram
    h := circonusllhist.New()

    //Insert value 123, three times
    h.RecordValues(123, 3)

    //Insert 1x10^1
    h.RecordIntScale(1,1)

    //Print the count of samples stored in the histogram
    fmt.Printf("%d\n",h.Count())

    //Print the sum of all samples
    fmt.Printf("%f\n",h.ApproxSum())
}
```
