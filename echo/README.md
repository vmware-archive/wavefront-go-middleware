# Middleware-echoweb
## Instrumentation

1. Create `tracer-config.yaml file`
2. Create `tracer-routesRegistration`
3. Call  `init` function and initialise tracer with config file path, routes registration file path, choice of `direct` or `proxy` sender in bool,and wavefront sender config.
4. Copy tracer-config.yaml, tracer-routesRegistration.yaml file to container in Dockerfile

## Example: tracer-config.yaml file:
```
cluster* : production
shard* : 1
application : devops-insight
service : devops-insight-api
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
	configFilePath := "filePath of tracer-config.yaml file"
	routesRegistrationFilePath := "filePath of tracer-routesRegistration.yaml file"
	
	//Initialising Global tracer
	err := InitTracer(configFilePath,routesRegistrationFilePath ,EchoWeb, true, wavefrontUrl, Key,flushIntervalSeconds)

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
	configFilePath := "filePath of tracer-config.yaml file"
	routesRegistrationFilePath := "filePath of tracer-routesRegistration.yaml file"
	
	//Initialising Global tracer
	err := InitTracer(configFilePath,routesRegistrationFilePath,EchoWeb,false, proxyHost, "",flushIntervalSeconds)

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
logger:= NewWfLogger(c)
```
  
*c - echo context
Same usage as default go logger instance given by log lib in go. Ex: logger.Println("Logging")
Logs are automatically injected with trace id, span id, and parent span id.
Logs are also sent to wavefront for each call*

**All the function exposed by standard Go Logger are exposed by custom logger with same usage**


##  Other tracing methods exposed 

### For manual start of span

```go
serverSpan,parentSpanId,err:= StartTraceSpan(c,operationName,tags)
```
  *c - echo context*
  
### For injecting headers in default http client request 
```go
httpRequest:= InjectTracerHttp(tracer,span,httpReq)
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
