# wavefront-go-middleware [![travis build status](https://travis-ci.com/wavefrontHQ/wavefront-go-middleware.svg?branch=master)](https://travis-ci.com/wavefrontHQ/wavefront-go-middleware) [![Go Report Card](https://goreportcard.com/badge/github.com/wavefrontHQ/wavefront-go-middleware)](https://goreportcard.com/report/github.com/wavefrontHQ/wavefront-go-middleware) [![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

This library provides wavefront-opentracing-middleware support for various Go web frameworks. The middleware handles the entire lifecycle of reporting spans with minimal code injection within APIs. It provides a custom logger abstracted on top of Go standard Logger which injects request-scoped unique trace identifiers into logs generated while servicing an Api request.

## Requirements
-   Go 1.10 or higher.

## Usage

### <a href="https://echo.labstack.com"><img height="20" src="https://cdn.labstack.com/images/echo-logo.svg"></a>

Import the `echo` package and instrument using the steps detailed in `wavefront-go-middleware/echo/README.md`
```go
import (
    wavefront "github.com/wavefronthq/wavefront-go-middleware/echo"
)
```

## License
[Apache 2.0 License](LICENSE).