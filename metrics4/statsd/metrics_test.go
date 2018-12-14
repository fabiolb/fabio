package statsd

import (
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

func TestLabeledCounters(t *testing.T) {
	counterMessage := "fabio_counter_code_200:1.000000|c\n"

	provider, err := NewProvider(addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		timer := time.NewTimer(5 * time.Second)
		<-timer.C
		t.Fatal("timeout")
	}()

	defer l.Close()

	provider.NewCounter("counter", "code").With("code", "200").Add(1)

	buffer := make([]byte, len(counterMessage))

	read, _, err := l.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if msg := string(buffer[:read]); msg != counterMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", msg, counterMessage)
	}
}

func TestLabeledTimers(t *testing.T) {
	timerMessage := "fabio_timer_code_200:0.500000|ms"

	provider, err := NewProvider(addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		timer := time.NewTimer(5 * time.Second)
		<-timer.C
		t.Fatal("timeout")
	}()

	defer l.Close()

	provider.NewTimer("timer", "code").With("code", "200").Observe(0.5)

	buffer := make([]byte, len(timerMessage))

	read, _, err := l.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if msg := string(buffer[:read]); msg != timerMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", msg, timerMessage)
	}
}

func TestLabeledGauges(t *testing.T) {
	gaugeMessage := "fabio_gauge_code_200:5.000000|g"

	provider, err := NewProvider(addr, time.Second)
	if err != nil {
		t.Fatal(err)
	}

	l, err := net.ListenPacket("udp", addr)
	if err != nil {
		t.Fatal(err)
	}

	go func() {
		timer := time.NewTimer(5 * time.Second)
		<-timer.C
		t.Fatal("timeout")
	}()

	defer l.Close()

	provider.NewGauge("gauge", "code").With("code", "200").Add(5)

	buffer := make([]byte, len(gaugeMessage))

	read, _, err := l.ReadFrom(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if msg := string(buffer[:read]); msg != gaugeMessage {
		t.Fatalf("Unexpected message:\nGot:\t%s\nExpected:\t%s\n", msg, gaugeMessage)
	}
}
