package trace

import (
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	mocktracer "github.com/opentracing/opentracing-go/mocktracer"
	"net/http"
	"testing"
)

func mimicTracerInject(req *http.Request) {
	// TODO maybe replace this will a call to opentracing.GlobalTracer().Inject()
	req.Header.Add("X-B3-TraceId", "1234562345678")
	req.Header.Add("X-B3-SpanId", "123456789")
	req.Header.Add("X-B3-ParentSpanId", "123456789")
	req.Header.Add("X-B3-Flags", "1")
}

// go test -v ./trace
func TestInjectHeaders(t *testing.T) {
	serviceName := "TESTING"

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Error("Error when creating new request.")
		t.Fail()
	}
	mimicTracerInject(req)

	mt := mocktracer.New()
	opentracing.SetGlobalTracer(mt)
	globalTracer := opentracing.GlobalTracer()

	var span opentracing.Span
	if globalTracer != nil {
		spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			span = globalTracer.StartSpan(serviceName)
		} else {
			span = globalTracer.StartSpan(serviceName, ext.RPCServerOption(spanCtx))
		}
	}

	InjectHeaders(span, req)

	if req.Header.Get("X-B3-Traceid") == "" {
		t.Error("Zipkin headers not set in request.")
		t.Fail()
	}
	if req.Header.Get("X-B3-Traceid") != "1234562345678" {
		t.Error("Zipkin headers do not match the values set.")
		t.Fail()
	}
}
