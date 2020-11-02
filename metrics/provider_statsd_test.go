package metrics

import (
	"bufio"
	"bytes"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/statsd"
	"io"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestStatsdProvider(t *testing.T) {
	prefix := "test-"
	s := statsd.New(prefix, log.NewNopLogger())
	provider := &StatsdProvider{S: s}
	for _, tst := range []struct {
		name         string
		metricname   string
		expectedname string
		labels       []string
		values       []string
		countval     float64
		gaugeval     float64
		histoval     float64
	}{
		{
			name:         "simpleMetrics",
			metricname:   "simple",
			expectedname: "simple",
			countval:     1,
			gaugeval:     2,
			histoval:     (time.Millisecond * 5).Seconds(),
		},
		{
			name:         "routeMetrics",
			metricname:   "route",
			labels:       []string{"service", "host", "path", "target"},
			values:       []string{"service", "foo", "host", "bar", "path", "/asdf", "target", "http://jkl.org:1234"},
			expectedname: "foo.bar./asdf.jkl_org_1234",
			countval:     20,
			gaugeval:     30,
			histoval:     (time.Millisecond * 50).Seconds(),
		},
		{
			name:         "codeMetrics",
			metricname:   "status",
			labels:       []string{"code"},
			values:       []string{"code", "200"},
			expectedname: "status.{type}.code.200",
			countval:     60,
			gaugeval:     70,
			histoval:     (time.Millisecond * 80).Seconds(),
		},
	} {
		t.Run(tst.name, func(t *testing.T) {
			cname := tst.metricname + ".count"
			gname := tst.metricname + ".gauge"
			hname := tst.metricname + ".histo"
			counter := provider.NewCounter(cname, tst.labels...)
			gauge := provider.NewGauge(gname, tst.labels...)
			histo := provider.NewHistogram(hname, tst.labels...)
			if len(tst.labels) > 0 {
				counter = counter.With(tst.values...)
				gauge = gauge.With(tst.values...)
				histo = histo.With(tst.values...)
			}
			counter.Add(tst.countval)
			gauge.Set(tst.gaugeval)
			histo.Observe(tst.histoval)
			var buff bytes.Buffer
			_, _ = provider.S.WriteTo(&buff)
			m := parseMetrics(&buff, prefix)
			// t.Logf("parsed metrics: %#v", m)

			for _, v := range []struct {
				n string
				v float64
			}{
				{"count", tst.countval},
				{"gauge", tst.gaugeval},
				{"histo", tst.histoval * 1000.0},
			} {
				var name string
				// have to do this little dance because route metrics
				// follow a special rule
				if strings.Contains(tst.expectedname, "{type}") {
					name = strings.ReplaceAll(tst.expectedname, "{type}", v.n)
				} else {
					name = tst.expectedname + "." + v.n
				}
				if se, ok := m[name]; ok {
					if se.value != v.v {
						t.Errorf("%s failed: expected: %.02f, got: %02f", name, v.v, se.value)
					}
				} else {
					t.Errorf("%s not found", v.n)
				}
			}
		})
	}

}

type statsdEntry struct {
	value  float64
	t      string
	sample float64
}

var re = regexp.MustCompile(`^([^:]+):([0-9\.]+)\|(ms|c|g)(?:\|@([0-9\.]+))?$`)

func parseMetrics(data io.Reader, prefix string) map[string]statsdEntry {
	reader := bufio.NewScanner(data)
	m := make(map[string]statsdEntry)
	for reader.Scan() {
		line := reader.Text()
		matches := re.FindStringSubmatch(line)
		if matches == nil {
			panic(line)
		}
		name := strings.TrimPrefix(matches[1], prefix)
		value, err := strconv.ParseFloat(matches[2], 64)
		if err != nil {
			panic(err.Error())
		}
		var sample float64
		if len(matches[4]) > 0 {
			sample, err = strconv.ParseFloat(matches[4], 64)
			if err != nil {
				panic(err.Error)
			}
		}

		m[name] = statsdEntry{
			value:  value,
			t:      matches[3],
			sample: sample,
		}
	}
	return m
}
