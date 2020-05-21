# Middleware-echoweb

## Instrumentation

1. Create `Cfg.yaml file`
2. Create `Routes.yaml file`
3. Initailise `Config` with choice of sender, absolute path of CfgFile and RoutesFile along with EchoWeb Server pointer.
3. Call  `init` function and pass the `Config` struct intialised above.
4. Copy tracer-config.yaml, tracer-routesRegistration.yaml file to container in Dockerfile

## Example: tracer-config.yaml file:
```
cluster* : prod
shard* : 1
application : wavefrontHQ
service : wavefront-middleware
source : VMware
CustomApplicationTags* :
	staticTags:
		customApplicationTagKey : customApplicationTagValue
	jwtClaims**:
		- username
rateSampler* : 10
durationSampler* : 60
```

**optional parameter*
*If no rate sampler is given it is taken as deterministic sampler ie all traces are sent to wavefront*

***Middleware automatically parses jwt token to add the given claims as tags in the span*

## Example: tracer-routesRegistration.yaml
```
routesRegistration:
		/tracingapi/test.GET:
				routeSpecificTag: route-tag
				routeSpecifcMetaTags:
					routeSpecifcMetaTagKey: routeSpecifcMetaTagValue
```
**Routes registration format:**
api-path.HTTP_METHOD


## init function:
### Using Direct Sender
```go
func init() {
	cfgFile := "filePath of Cfg.yaml file"
	routesFile := "filePath of Routes.yaml file"

	config := new(Config)
	config.CfgFile = cfgFile
	config.RoutesFile= routesFile
	config.echoWeb = *echo.Echo
	config.DirectCfg = senders.DirectConfiguration
	
	//Initialising Global tracer
	err := InitTracer(config)

	if err != nil {
		/*
		Handle Error
		*/
	}
}
```

### Using In-Direct Sender
```go
func init() {
	cfgFile := "filePath of Cfg.yaml file"
	routesFile := "filePath of Routes.yaml file"

	config := new(Config)
	config.CfgFile = cfgFile
	config.RoutesFile= routesFile
	config.echoWeb = *echo.Echo
	config.DirectCfg = senders.ProxyConfiguration
	
	//Initialising Global tracer
	err := InitTracer(config)

	if err != nil {
		/*
		Handle Error
		*/
	}
}
```
## Cross process context propagation
```go
headers:= GetTracingHeadersToInject(c)
```
*c - echo context*

 `GetTracingHeadersToInject` func returns headers which can be added to the call when making cross service calls. `Context` is propagated using headers which helps in stiching trace 

## Adding dynamic tags to Span
To add metadata derived during servicing the request
```go
AddDynamicTags(c,tags)
```
*c - echo context
tags - key value pairs of strings*

## Contextual Logger
```go
logger:= NewSpanLogger(c)
```
  
*c - echo context
Same usage as default go logger instance given by log lib in go. Ex: logger.Println("Logging")
Logs are automatically injected span info including trace id, span id, and parent span id.
Logs are sent to wavefront for each call* 

**All the function exposed by standard Go Logger are exposed by custom logger with same usage**


##  Other tracing methods exposed 

### For manual start of span

```go
serverSpan,parentSpanId,err:= StartTraceSpan(c,operationName,tags)
```
  *c - echo context*
  
### For injecting headers in default http client request 
```go
httpRequest:= InjectTracerHTTP(tracer,span,httpReq)
```

### Returns headers to be injected from span. To be used when manually starting span within the process
```go
tags:= GetTracingHeadersToInjectFromSpan(tracer,span)
```
  *c - echo context*


### Returns headers to be injected from echo context. To be used when using middleware
```go
GetTracingHeadersToInjectFromContext(c)
```
