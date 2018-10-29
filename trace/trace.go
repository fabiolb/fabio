package trace

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/fabiolb/fabio/config"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
)

func InjectHeaders(span opentracing.Span, req *http.Request) {
	// Inject span data into the request headers
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

func CreateCollector(collectorType, connectString, topic string) zipkin.Collector {
	var collector zipkin.Collector
	var err error

	switch collectorType {
	case "http":
		collector, err = zipkin.NewHTTPCollector(connectString)
	case "kafka":
		// TODO set logger?
		kafkaHosts := strings.Split(connectString, ",")
		collector, err = zipkin.NewKafkaCollector(
			kafkaHosts,
			zipkin.KafkaTopic(topic),
		)
	default:
		err = fmt.Errorf("Unknown collector type.")
	}

	if err != nil {
		log.Fatalf("Unable to create Zipkin %s collector: %v", collectorType, err)
	}

	return collector
}

func CreateTracer(recorder zipkin.SpanRecorder, samplerRate float64) opentracing.Tracer {
	tracer, err := zipkin.NewTracer(
		recorder,
		zipkin.WithSampler(zipkin.NewBoundarySampler(samplerRate, 1)),
		zipkin.ClientServerSameSpan(false),
		zipkin.TraceID128Bit(true),
	)

	if err != nil {
		log.Printf("Unable to create Zipkin tracer: %+v", err)
		os.Exit(-1)
	}

	return tracer
}

func CreateSpan(r *http.Request, serviceName string) opentracing.Span {
	globalTracer := opentracing.GlobalTracer()

	// If headers contain trace data, create child span from parent; else, create root span
	var span opentracing.Span
	if globalTracer != nil {
		spanCtx, err := globalTracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		if err != nil {
			span = globalTracer.StartSpan(serviceName)
		} else {
			span = globalTracer.StartSpan(serviceName, ext.RPCServerOption(spanCtx))
		}
	}

	return span // caller must defer span.finish()
}

// InitializeTracer initializes OpenTracing support if Tracing.TracingEnabled
// is set in the config.
func InitializeTracer(traceConfig *config.Tracing) {
	if !traceConfig.TracingEnabled {
		return
	}

	log.Printf("Tracing initializing - type: %s, connection string: %s, service name: %s, topic: %s, samplerRate: %v",
		traceConfig.CollectorType, traceConfig.ConnectString, traceConfig.ServiceName, traceConfig.Topic, traceConfig.SamplerRate)

	// Create a new Zipkin Collector, Recorder, and Tracer
	collector := CreateCollector(traceConfig.CollectorType, traceConfig.ConnectString, traceConfig.Topic)
	recorder := zipkin.NewRecorder(collector, false, traceConfig.SpanHost, traceConfig.ServiceName)
	tracer := CreateTracer(recorder, traceConfig.SamplerRate)

	// Set the Zipkin Tracer created above to the GlobalTracer
	opentracing.SetGlobalTracer(tracer)
}
