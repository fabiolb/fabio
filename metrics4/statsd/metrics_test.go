package statsd

import (
	"github.com/fabiolb/fabio/metrics4"
	"net"
	"testing"
	"time"
)

const addr = ":9876"

// It shouldn't panic after creating several metrics with the same name
func TestIdenticalNamesForCounters(t *testing.T) {
	metricName := "metric"
	provider, err := NewProvider("addr", 5 * time.Second)

	if err != nil {
		t.Error(err)
	}

	counter := provider.NewCounter(metricName)
	counter.Add(1)
	counter = provider.NewCounter(metricName)
	counter.Add(1)
}

func createProvider(t *testing.T, addr string, interval time.Duration) metrics4.Provider {
	provider, err := NewProvider(addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	return provider
}

func createUdpConnection(t *testing.T, addr string) net.PacketConn {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatal(err)
	}
	return conn
}

func readStringFromUdpConnection(t *testing.T, conn net.PacketConn, length int) string {
	buffer := make([]byte, length)
	read, _, err := conn.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}
	return string(buffer[:read])
}

func startTimeout(t *testing.T, duration time.Duration) {
	go func() {
		timer := time.NewTimer(duration)
		<-timer.C
		t.Fatal("timeout")
	}()
}

func TestLabeledCounters(t *testing.T) {
	counterMessage := "fabio_counter_code_200:1.000000|c\n"

	provider := createProvider(t, addr, time.Second)
	defer provider.Close()

	conn := createUdpConnection(t, addr)
	defer conn.Close()

	startTimeout(t, 5 * time.Second)

	provider.NewCounter("counter", "code").With("code", "200").Add(1)

	message := readStringFromUdpConnection(t, conn, len(counterMessage))

	if message != counterMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", message, counterMessage)
	}
}

func TestLabeledTimers(t *testing.T) {
	timerMessage := "fabio_timer_code_200:0.500000|ms"

	provider := createProvider(t, addr, time.Second)
	defer provider.Close()

	conn := createUdpConnection(t, addr)
	defer conn.Close()

	startTimeout(t, 5 * time.Second)

	provider.NewTimer("timer", "code").With("code", "200").Observe(0.5)

	message := readStringFromUdpConnection(t, conn, len(timerMessage))

	if message != timerMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", message, timerMessage)
	}
}

func TestLabeledGauges(t *testing.T) {
	gaugeMessage := "fabio_gauge_code_200:5.000000|g"

	provider := createProvider(t, addr, time.Second)
	defer provider.Close()

	conn := createUdpConnection(t, addr)
	defer conn.Close()

	startTimeout(t, 5 * time.Second)

	provider.NewGauge("gauge", "code").With("code", "200").Add(5)

	message := readStringFromUdpConnection(t, conn, len(gaugeMessage))

	if message != gaugeMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", message, gaugeMessage)
	}
}
