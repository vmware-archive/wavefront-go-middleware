package echo

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-opentracing-sdk-go/tracer"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func TestGetTracingHeadersToInjectFromSpan(t *testing.T) {
	reporter := tracer.NewInMemoryReporter()
	testTracer := tracer.New(reporter)
	spanContext := tracer.SpanContext{
		TraceID: "3871de7e09c53ae8",
		SpanID:  "spanId",
		Sampled: nil,
		Baggage: nil,
	}

	sp := testTracer.StartSpan("test", opentracing.ChildOf(spanContext))
	headers := GetTracingHeadersToInjectFromSpan(testTracer, sp)
	chilSpanContext := sp.Context().(tracer.SpanContext)

	assert.Equal(t, headers["Wf-Ot-Traceid"], chilSpanContext.TraceID)
	assert.Equal(t, headers["Wf-Ot-Spanid"], chilSpanContext.SpanID)

}

func TestGetTracingHeadersToInjectFromContext(t *testing.T) {
	e := echo.New()
	directCfg := &senders.DirectConfiguration{
		Server:               "http://localhost:" + "8080",
		Token:                "DUMMY_TOKEN",
		BatchSize:            10000,
		MaxBufferSize:        50000,
		FlushIntervalSeconds: 1,
	}

	absCfgFilePath, err := filepath.Abs("./Cfg_test.yaml")
	if err != nil {
		t.Error("Error finding CfgFilePath", err)
	}

	absRoutesFilePath, err := filepath.Abs("./Routes_test.yaml")
	if err != nil {
		t.Error("Error finding RoutesFilePath", err)
	}

	cfg := Config{
		CfgFile:    absCfgFilePath,
		RoutesFile: absRoutesFilePath,
		EchoWeb:    e,
		DirectCfg:  directCfg,
	}

	err = InitTracer(cfg)
	if err != nil {
		t.Error("Failed Initalising Tracer", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/")
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	TracingHandler(handler)(c)

	spanPointerInter := c.Get("spanPointer")
	spanPointer := (spanPointerInter).(*opentracing.Span)
	propagationHeaders := GetTracingHeadersToInjectFromContext(c)
	chilSpanContext := (*spanPointer).Context().(tracer.SpanContext)
	assert.Equal(t, propagationHeaders["Wf-Ot-Traceid"], chilSpanContext.TraceID)
	assert.Equal(t, propagationHeaders["Wf-Ot-Spanid"], chilSpanContext.SpanID)

}

func TestStartTraceSpan(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	span, ParentId := StartTraceSpan(c, "", nil)

	assert.NotNil(t, span)
	assert.Equal(t, ParentId, "")

}
