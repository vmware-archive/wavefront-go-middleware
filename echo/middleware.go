package echo

import (
	"fmt"
	"log"

	"github.com/gookit/color"
	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/wavefronthq/wavefront-opentracing-sdk-go/tracer"
)

//AddDynamicTags injects the dyanmic tags to the span
func AddDynamicTags(context *echo.Context, tags map[string]interface{}) {
	(*context).Set("dynamicTags", tags)
}

//TracingHandler custom echoWeb middleware.
//Enables tracing for the routes defined in tracer-config.go.
//Injects respective tags for each route defined in the span of each trace.
func TracingHandler(next echo.HandlerFunc) echo.HandlerFunc {
	return func(context echo.Context) error {
		//Mapping path and request type to Http request to get corresponding tags
		if routeToTagMapValue, ok := globalTracerConfig.RouteToTagsMap[context.Path()+"."+context.Request().Method]; ok {
			span, parentSpanID := StartTraceSpan(context, routeToTagMapValue.OperationName, routeToTagMapValue.Metatags)

			/*EXTRACTING TRACID AND SPANID FROM SPAN
			  CREATING THE IDENTIFIER TO INJECT IN CONTEXT*/
			logPrefix := ""
			sc, ok := span.Context().(tracer.SpanContext)
			if ok {
				//If it is null both parent and child will have same span id. Adding parent.span.id as tag
				if parentSpanID != "" {
					logPrefix = "traceId:" + sc.TraceID + " " + "spanId:" + sc.SpanID + " " + "parentId:" + parentSpanID + " "
					span.SetTag("parentId", parentSpanID)
				} else {
					logPrefix = "traceId:" + sc.TraceID + " " + "spanId:" + sc.SpanID + " " + "parentId:" + sc.SpanID + " "
					span.SetTag("parentId", sc.SpanID)
				}

				//Span reference in context
				context.Set("spanPointer", &span)
			}

			//Injecting the prefix in echo context
			context.Set("tracePrefix", logPrefix)

			//Injecting propagation headers in context
			propagationHeaders := GetTracingHeadersToInjectFromSpan(Tracer, span)
			context.Set("propagationHeaders", propagationHeaders)

			//Executing function after the response is written
			context.Response().After(func() {
				populateTagsForTrace(context, &span)
				//Finishing the span
				span.Finish()
			})
		} else {
			log.Println(color.Warn.Sprint("Wavefront Middleware... Tracing not configured for this route"))
		}
		return next(context)
	}
}

//populateDefaultTagsForTrace extracts the required information from context and adds them as tags to the trace
func populateTagsForTrace(context echo.Context, span *opentracing.Span) {
	//Extracting Http info to populate as default tags
	requestMethod := context.Request().Method
	requestURL := context.Request().Host + context.Request().RequestURI
	requestStatus := context.Response().Status
	userAgent := context.Request().UserAgent()
	host := context.Request().Host

	//Populating tags with http info
	(*span).SetTag("http.url", requestURL)
	(*span).SetTag("http.method", requestMethod)
	(*span).SetTag("http.status_code", requestStatus)
	(*span).SetTag("http.user_agent", userAgent)
	(*span).SetTag("http.hostname", host)

	if requestStatus > 399 {
		(*span).SetTag("error", "true")
	}

	//Populating dynamically added tags
	populateDynamicTagsForTrace(context, span)

	//Populating JWT tags
	if len(globalTracerConfig.CustomApplicationTags.JwtClaims) > 0 {
		populateJwtTagsByClaims(context, span)
	}
}

//populateDynamicTagsForTrace extracts the dynamically added tags from context and adds them as tags to the trace
func populateDynamicTagsForTrace(context echo.Context, span *opentracing.Span) {
	//Extracing the tags from echo context
	dynamicTags := context.Get("dynamicTags")
	tags, ok := dynamicTags.(map[string]interface{})
	if !ok {
		return
	}

	//Populating the tags
	for key, value := range tags {
		(*span).SetTag(key, value)
	}
}

//populateJwtTagsByClaims
func populateJwtTagsByClaims(context echo.Context, span *opentracing.Span) {
	jwtClaimsFields := globalTracerConfig.CustomApplicationTags.JwtClaims

	// Fetch all claims from JWT
	claims, err := parseJwtClaimsFromHeaders(context)
	if err != nil {
		return
	}

	// Fetch claim value. Throw error if username
	// claim is not present
	for _, claim := range jwtClaimsFields {
		claimValue := fmt.Sprintf("%v", claims[claim])
		if claimValue != "" {
			(*span).SetTag("jwt."+claim, claimValue)
		}
	}
}
