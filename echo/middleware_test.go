package echo

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
	"github.com/wavefronthq/wavefront-sdk-go/senders"
)

func TestMiddlewareContextInjection(t *testing.T) {
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
		t.Error("Failed Initialising Tracer", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/")

	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "test")
	}

	TracingHandler(handler)(c)

	spanPointer := c.Get("spanPointer")
	assert.NotNil(t, spanPointer)
	assert.IsType(t, (spanPointer).(*opentracing.Span), spanPointer)

	tracePrefix := c.Get("tracePrefix")
	assert.NotNil(t, tracePrefix)

	propagationHeaders := c.Get("propagationHeaders")
	assert.NotNil(t, propagationHeaders)
}

func TestAddDynamicTag(t *testing.T) {
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
		t.Error("Failed Initialising Tracer", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/")
	handler := func(c echo.Context) error {
		tags := make(map[string]interface{}, 1)
		tags["test"] = "middlewareTestTag"
		AddDynamicTags(&c, tags)
		return c.String(http.StatusOK, "test")
	}

	TracingHandler(handler)(c)

	dynamicTags := c.Get("dynamicTags")
	assert.NotNil(t, dynamicTags)
	assert.Equal(t, dynamicTags.(map[string]interface{})["test"], "middlewareTestTag")
}
