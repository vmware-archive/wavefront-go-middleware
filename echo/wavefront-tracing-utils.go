package echo

import (
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
func InitTracer(configFilePath string, routesRegistrationFilePath string, echoWeb *echo.Echo, sendDirectly bool, server string, token string, flushIntervalSeconds int) error {

	//Constructing globalTracerConfig from config file
	tracerConfig, err := readConfigFromYaml(configFilePath)
	if err != nil {
		return err
	}
	//Registering routes
	tracerConfig, err = readConfigFromYaml(routesRegistrationFilePath)
	if err != nil {
		return err
	}

	var sender senders.Sender
	if sendDirectly {
		// Create the direct sender
		sender, err = getWavefrontDirectSender(server, token, flushIntervalSeconds)
		if err != nil {
			return err
		}
	} else {
		// Create the proxy sender
		sender, err = getWavefrontProxySender(server, flushIntervalSeconds)
		if err != nil {
			return err
		}
	}
	appTags := wfApplication.New(tracerConfig.Application, tracerConfig.Service)
	if tracerConfig != nil {
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
	}
	wfReporter := reporter.New(sender, appTags, reporter.Source(tracerConfig.Source))
	clReporter := reporter.NewConsoleSpanReporter(tracerConfig.Source)
	reporter := reporter.NewCompositeSpanReporter(wfReporter, clReporter)
	var durationSampler tracer.DurationSampler
	var rateSampler tracer.RateSampler

	//Initialising tracer with RateSampler and DurationSampler if given in conifg
	//If values for RateSampler and DurationSampler is not given, it initalises standard tracer with default values
	if tracerConfig.DurationSampler > 0 && tracerConfig.RateSampler > 0 {
		durationSampler = tracer.DurationSampler{Duration: time.Duration(tracerConfig.DurationSampler) * time.Second}
		rateSampler = tracer.RateSampler{Rate: tracerConfig.RateSampler}

		Tracer = wfTracer.New(reporter, wfTracer.WithSampler(rateSampler), wfTracer.WithSampler(durationSampler))
	} else if tracerConfig.RateSampler > 0 {
		rateSampler = tracer.RateSampler{Rate: tracerConfig.RateSampler}

		Tracer = wfTracer.New(reporter, wfTracer.WithSampler(rateSampler))
	} else if tracerConfig.DurationSampler > 0 {
		durationSampler = tracer.DurationSampler{Duration: time.Duration(tracerConfig.DurationSampler) * time.Second}

		Tracer = wfTracer.New(reporter, wfTracer.WithSampler(durationSampler))
	} else {
		Tracer = wfTracer.New(reporter)
	}

	//Instantiating Global Tracer
	opentracing.InitGlobalTracer(Tracer)

	//Enabling Middleware
	echoWeb.Use(middlewareTracing)
	log.Println("Tracer Initialized...")
	return nil
}

//StartTraceSpan looks for existing context from the headers of the request.
//If context is found, it starts a child span to the parent span,
//Else it starts a root span.
//Tags are added to the respective spans created.
//If parent span exists parent span id is returned along with child span otherwise root span and "" is returned
func StartTraceSpan(c echo.Context, appSpecificOperationName string, tags map[string]string) (opentracing.Span, string, error) {
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

	return serverSpan, parentSpanID, nil
}

//InjectTracerHTTP injects the current span into the tracer.
//Adds the context of the current span to the headers of the Http request
//to the previous span of the tracer.
func InjectTracerHTTP(tracer opentracing.Tracer, serverSpan opentracing.Span, httpReq *http.Request) *http.Request {
	tracer.Inject(
		serverSpan.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(httpReq.Header))
	return httpReq
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
	headersToBeInjected := make(map[string]string)
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
	headersToBeInjected := context.Get("propagationHeaders").(map[string]string)
	return headersToBeInjected
}
