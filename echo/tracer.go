package echo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	reporter "github.com/wavefronthq/wavefront-opentracing-sdk-go/reporter"
	"github.com/wavefronthq/wavefront-opentracing-sdk-go/tracer"
	wfTracer "github.com/wavefronthq/wavefront-opentracing-sdk-go/tracer"
	wfApplication "github.com/wavefronthq/wavefront-sdk-go/application"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

// Tracer Global Tracer
var Tracer opentracing.Tracer
var globalTracerConfig *tracerConfig

// InitTracer initialize tracer object, initialize the GlobalTracer with this newly created,
// tracer object.
// Call utils/wavefront.Tracer to access the tracer object
//For proxy sender send token as empty string
func InitTracer(cfg Config) error {

	//Constructing globalTracerConfig from config file
	tracerConfig, err := readConfigFromYaml(cfg.CfgFile)
	if err != nil {
		return fmt.Errorf("error reading config file: %w", err)
	}
	//Registering routes
	tracerConfig, err = readConfigFromYaml(cfg.RoutesFile)
	if err != nil {
		return fmt.Errorf("error reading routes file: %w", err)
	}

	var sender senders.Sender
	if cfg.DirectCfg != nil {
		// Create the direct sender
		sender, err = senders.NewDirectSender(cfg.DirectCfg)
		if err != nil {
			return err
		}
	} else if cfg.ProxyCfg != nil {
		// Create the proxy sender
		sender, err = senders.NewProxySender(cfg.ProxyCfg)
		if err != nil {
			return err
		}
	} else {
		return errors.New("wavefront middleware: one of Direct sender or Proxy sender configuration is required")
	}

	appTags := wfApplication.New(tracerConfig.Application, tracerConfig.Service)
	if len(tracerConfig.Cluster) > 0 {
		appTags.Cluster = tracerConfig.Cluster
	}
	if len(tracerConfig.Shard) > 0 {
		appTags.Shard = tracerConfig.Shard
	}
	if tracerConfig.CustomApplicationTags.StaticTags != nil && len(tracerConfig.CustomApplicationTags.StaticTags) > 0 {
		for k, v := range tracerConfig.CustomApplicationTags.StaticTags {
			appTags.CustomTags[k] = v
		}
	}

	wfReporter := reporter.New(sender, appTags, reporter.Source(tracerConfig.Source))

	//Initialising tracer with RateSampler and DurationSampler if given in conifg
	//If values for RateSampler and DurationSampler is not given, it initalises standard tracer with default values
	var samplers []wfTracer.Option
	if tracerConfig.DurationSampler > 0 {
		samplers = append(samplers, wfTracer.WithSampler(tracer.DurationSampler{Duration: time.Duration(tracerConfig.DurationSampler) * time.Second}))
	}
	if tracerConfig.RateSampler > 0 {
		samplers = append(samplers, wfTracer.WithSampler(tracer.RateSampler{Rate: tracerConfig.RateSampler}))
	}
	Tracer = wfTracer.New(wfReporter, samplers...)

	//Instantiating Global Tracer
	opentracing.SetGlobalTracer(Tracer)

	//Enabling Middleware
	cfg.EchoWeb.Use(TracingHandler)

	log.Println("wavefront middleware: tracer initialized")

	return nil
}

//StartTraceSpan looks for existing context from the headers of the request.
//If context is found, it starts a child span to the parent span,
//Else it starts a root span.
//Tags are added to the respective spans created.
//If parent span exists parent span id is returned along with child span otherwise root span and "" is returned
func StartTraceSpan(c echo.Context, appSpecificOperationName string, tags map[string]string) (opentracing.Span, string) {
	var serverSpan opentracing.Span
	var parentSpanID string
	wireContext, err := opentracing.GlobalTracer().Extract(
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(c.Request().Header))
	if err != nil {
		log.Println(err)
	}

	//Adding parent span Id if exists
	sc, ok := wireContext.(tracer.SpanContext)
	if ok {
		parentSpanID = sc.SpanID
	}

	// Create the span referring to the RPC client if available.
	// If wireContext == nil, a root span will be created.
	serverSpan = opentracing.StartSpan(
		appSpecificOperationName,
		ext.RPCServerOption(wireContext))

	//Populating Tags
	for key, value := range tags {
		if value != "" {
			serverSpan.SetTag(key, value)
		}
	}

	return serverSpan, parentSpanID
}

//InjectTracerHTTP injects the current span into the trace.
//By Adding the context of the current span to the headers of the Http request.
func InjectTracerHTTP(tracer opentracing.Tracer, serverSpan opentracing.Span, httpReq *http.Request) error {
	err := tracer.Inject(
		serverSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	return err
}

//GetTracingHeadersToInjectFromSpan returns the headers
//which Should be added to the request to inject from
//the current span.
//To be used when manually starting span inside process
func GetTracingHeadersToInjectFromSpan(tracer opentracing.Tracer, serverSpan opentracing.Span) map[string]string {
	httpReq, _ := http.NewRequest("GET", "serviceUrl", nil)
	tracer.Inject(
		serverSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	headersToBeInjected := make(map[string]string, len(httpReq.Header))
	for key, value := range httpReq.Header {
		headersToBeInjected[key] = value[0]
	}
	return headersToBeInjected
}

//GetTracingHeadersToInjectFromContext returns the headers
//which Should be added to the request to inject from
//the echo context
//To be used when using middleware
func GetTracingHeadersToInjectFromContext(context echo.Context) map[string]string {
	return context.Get("propagationHeaders").(map[string]string)
}
