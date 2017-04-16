package riemann

import (
	"fmt"
	"github.com/amir/raidman"
	"github.com/cenkalti/backoff"
	"github.com/rcrowley/go-metrics"
	"log"
	"os"
	"path"
	"time"
)

// Send all reported metrics to connected Riemann client
func ReportOnce(r metrics.Registry, c *raidman.Client) error {
	events := []*raidman.Event{}

	r.Each(func(name string, i interface{}) {
		switch metric := i.(type) {
		case metrics.Gauge:
			e := &raidman.Event{
				Service: name,
				Metric:  metric.Value(),
			}
			events = append(events, e)

		case metrics.GaugeFloat64:
			e := &raidman.Event{
				Service: name,
				Metric:  metric.Value(),
			}
			events = append(events, e)

		case metrics.Counter:
			e := &raidman.Event{
				Service: metricName(name),
				Metric:  int(metric.Count()),
			}
			events = append(events, e)

		case metrics.Meter:
			e := meterEvents(name, metric.Snapshot())
			copy(events, e)

		case metrics.Histogram:
			e := histogramEvents(name, metric.Snapshot())
			copy(events, e)

		case metrics.Timer:
			e := timerEvents(name, metric.Snapshot())
			copy(events, e)
		}
	})

	for _, e := range events {
		err := c.Send(e)
		if err != nil {
			log.Println("error sending riemann metric.", err)
			return err
		}
	}

	return nil
}

// Opens a connection to Riemann and repeatedly sends events
// for all metrics
func Report(r metrics.Registry, d time.Duration, riemannHost string) {
	var c = RiemannConnect(riemannHost)

	for _ = range time.Tick(d) {
		err := ReportOnce(r, c)

		if err != nil {
			log.Println("reconnecting to riemann")
			c.Close()
			c = RiemannConnect(riemannHost)
		}
	}
}

// Establishes a Riemann connection, will block (and retry) until
// it can successfully establish a connection.
func RiemannConnect(host string) *raidman.Client {
	connChannel := make(chan *raidman.Client)

	go func() {
		connect := func() error {
			c, err := raidman.Dial("tcp", host)
			if err != nil {
				log.Println("Error connecting to Riemann, will retry.", err)
				return err
			} else {
				log.Println("connected to riemann server", host)
				connChannel <- c
				return nil
			}
		}

		policy := &backoff.ConstantBackOff{time.Second * 5}
		backoff.Retry(connect, policy)
	}()

	return <-connChannel
}

func metricName(name string) string {
	return fmt.Sprintf("%s %s", path.Base(os.Args[0]), name)
}

func meterEvents(name string, metric metrics.Meter) []*raidman.Event {
	return []*raidman.Event{
		event(name, "count", int(metric.Count())),
		event(name, "mean", metric.RateMean()),
		event(name, "one-minute", metric.Rate1()),
		event(name, "five-minute", metric.Rate5()),
		event(name, "fifteen-minute", metric.Rate15()),
	}
}

func timerEvents(name string, metric metrics.Timer) []*raidman.Event {
	events := []*raidman.Event{
		event(name, "count", int(metric.Count())),
		event(name, "min", int(metric.Min())),
		event(name, "max", int(metric.Max())),
		event(name, "mean", metric.Mean()),
		event(name, "std-dev", metric.StdDev()),
		event(name, "one-minute", metric.Rate1()),
		event(name, "five-minute", metric.Rate5()),
		event(name, "fifteen-minute", metric.Rate15()),
	}
	percentiles := []float64{0.75, 0.95, 0.99, 0.999}
	percentileVals := metric.Percentiles(percentiles)
	for i, p := range percentiles {
		e := event(name, fmt.Sprintf("percentile-%.3f", p), percentileVals[i])
		events = append(events, e)
	}
	return events
}

func histogramEvents(name string, metric metrics.Histogram) []*raidman.Event {
	events := []*raidman.Event{
		event(name, "count", int(metric.Count())),
		event(name, "min", int(metric.Min())),
		event(name, "max", int(metric.Max())),
		event(name, "mean", metric.Mean()),
		event(name, "std-dev", metric.StdDev()),
	}

	percentiles := []float64{0.75, 0.95, 0.99, 0.999}
	percentileVals := metric.Percentiles(percentiles)
	for i, p := range percentiles {
		e := event(name, fmt.Sprintf("percentile-%.3f", p), percentileVals[i])
		events = append(events, e)
	}
	return events
}

func event(name string, measure string, val interface{}) *raidman.Event {
	return &raidman.Event{
		Host:    "",
		Service: metricName(fmt.Sprintf("%s.%s", name, measure)),
		Metric:  val,
	}
}
