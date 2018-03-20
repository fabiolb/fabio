package trace

import (
	"log"
	"net/http"
	"os"
	"strings"

	opentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go-opentracing"
	"github.com/opentracing/opentracing-go/ext"
)

func InjectHeaders(span opentracing.Span, req *http.Request) {
	// Inject span data into the request headers
	opentracing.GlobalTracer().Inject(
		span.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(req.Header),
	)
}

func CreateCollector(collectorType string, connectString string, topic string) zipkin.Collector {
	var collector zipkin.Collector
	var err error

	if collectorType == "http" {
		collector, err = zipkin.NewHTTPCollector(connectString)
	} else if collectorType == "kafka" {
		// TODO set logger?
		kafkaHosts := strings.Split(connectString, ",")
		collector, err = zipkin.NewKafkaCollector(
			kafkaHosts,
			zipkin.KafkaTopic(topic),
		)
	}

	if err != nil {
		log.Printf("Unable to create Zipkin %s collector: %+v", collectorType, err)
		os.Exit(-1)
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

func InitializeTracer(collectorType string, connectString string, serviceName string, topic string, samplerRate float64, addressPort string) {
	log.Printf("Tracing initializing - type: %s, connection string: %s, service name: %s, topic: %s, samplerRate: %v", collectorType, connectString, serviceName, topic, samplerRate)

	// Create a new Zipkin Collector, Recorder, and Tracer
	collector := CreateCollector(collectorType, connectString, topic)
	recorder := zipkin.NewRecorder(collector, false, addressPort, serviceName)
	tracer := CreateTracer(recorder, samplerRate)

	// Set the Zipkin Tracer created above to the GlobalTracer
	opentracing.SetGlobalTracer(tracer)

	log.Printf("\n\nTRACER: %v\n\n", tracer)
	log.Printf("\n\nCOLLECTOR: %v\n\n", collector)
	log.Printf("\n\nRECORDER: %v\n\n", recorder)
}