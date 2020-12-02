package metrics

import (
	"bytes"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/dogstatsd"
	"reflect"
	"testing"
	"time"
)

func TestDogstatsdProvider(t *testing.T) {
	prefix := "test-"
	d := dogstatsd.New(prefix, log.NewNopLogger())
	provider := &DogstatsdProvider{D: d}
	for _, tst := range []struct {
		name     string
		labels   []string
		values   []string
		prefix   string
		countval float64
		gaugeval float64
		histoval float64
	}{
		{
			name:     "simpleTest",
			labels:   []string{"service", "host", "path", "target", "other"},
			values:   []string{"service", "foo", "host", "bar", "path", "/asdf", "target", "http://jkl.org:1234", "other", "trailer"},
			prefix:   "tst",
			countval: 20,
			gaugeval: 30,
			histoval: (time.Microsecond * 50).Seconds(),
		},
	} {
		t.Run(tst.name, func(t *testing.T) {
			counter := provider.NewCounter(tst.prefix+".counter", tst.labels...)
			gauge := provider.NewGauge(tst.prefix+".gauge", tst.labels...)
			histo := provider.NewHistogram(tst.prefix+".histogram", tst.labels...)
			if len(tst.labels) > 0 {
				counter = counter.With(tst.values...)
				gauge = gauge.With(tst.values...)
				histo = histo.With(tst.values...)
			}
			counter.Add(tst.countval)
			gauge.Set(tst.gaugeval)
			histo.Observe(tst.histoval)
			var buff bytes.Buffer
			_, _ = provider.D.WriteTo(&buff)
			m := parseStatsdMetrics(&buff, prefix)
			for _, v := range []struct {
				n string
				v float64
			}{
				{tst.prefix + ".counter", tst.countval},
				{tst.prefix + ".gauge", tst.gaugeval},
				{tst.prefix + ".histogram", tst.histoval},
			} {
				if se, ok := m[v.n]; ok {
					if se.value != v.v {
						t.Errorf("%s failed: expected: %.02f, got %02f", v.n, v.v, se.value)
					}
					if len(tst.values) > 0 && !reflect.DeepEqual(se.tags, tst.values) {
						t.Errorf("tags did not survive round trip parsing")
					}
				} else {
					t.Errorf("%s not found", v.n)
				}
			}
		})
	}
}
