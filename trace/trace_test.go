package trace

import (
	"net/http"
	"testing"

	opentracing "github.com/opentracing/opentracing-go"
	mocktracer "github.com/opentracing/opentracing-go/mocktracer"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	zipkintypes "github.com/openzipkin/zipkin-go-opentracing/types"
)

const testServiceName = "TEST-SERVICE"

func TestInjectHeaders(t *testing.T) {
	mt := mocktracer.New()
	opentracing.SetGlobalTracer(mt)
	globalTracer := opentracing.GlobalTracer()
	span := globalTracer.StartSpan(testServiceName)

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	InjectHeaders(span, req)

	if req.Header.Get("Mockpfx-Ids-Traceid") == "" {
		t.Error("Inject did not set the Traceid in the request.")
		t.Fail()
	}
	if req.Header.Get("Mockpfx-Ids-Spanid") == "" {
		t.Error("Inject did not set the Spanid in the request.")
		t.Fail()
	}
}

func TestInjectHeadersWithParentSpan(t *testing.T) {
	parentSpanId := uint64(12345)
	parentSpanContext := zipkin.SpanContext{
		SpanID:  parentSpanId,
		TraceID: zipkintypes.TraceID{High: uint64(1234), Low: uint64(4321)},
	}

	tracer, _ := zipkin.NewTracer(nil)
	opentracing.SetGlobalTracer(tracer)
	globalTracer := opentracing.GlobalTracer()
	childSpan := globalTracer.StartSpan(testServiceName+"-CHILD", opentracing.ChildOf(parentSpanContext))

	req, _ := http.NewRequest("GET", "http://example.com", nil)
	InjectHeaders(childSpan, req)

	if req.Header.Get("X-B3-Traceid") == "" {
		t.Error("Inject did not set the Traceid in the request.")
		t.Fail()
	}
	if req.Header.Get("X-B3-Spanid") == "" {
		t.Error("Inject did not set the Spanid in the request.")
		t.Fail()
	}
	if req.Header.Get("X-B3-Parentspanid") != "0000000000003039" {
		t.Error("Inject did not set the correct Parentspanid in the request.")
		t.Fail()
	}
	if req.Header.Get("x-B3-Traceid") != "00000000000004d200000000000010e1" {
		t.Error("Inject did not reuse the Traceid from the parent span")
		t.Fail()
	}
}
