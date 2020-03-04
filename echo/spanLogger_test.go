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

func TestContextIdentifierLogPrefixInjection(t *testing.T) {
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

	lgger := NewSpanLogger(c)
	prefix := lgger.Prefix()
	spanPointerInter := c.Get("spanPointer")
	spanPointer := (spanPointerInter).(*opentracing.Span)
	sc := (*spanPointer).Context().(tracer.SpanContext)
	assert.Equal(t, prefix, "traceId:"+sc.TraceID+" "+"spanId:"+sc.SpanID+" "+"parentId:"+sc.SpanID+" ")

}
