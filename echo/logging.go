package echo

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/labstack/echo"
	"github.com/opentracing/opentracing-go"
)

//SpanLogger custom implementation of Go standard Logger
//Exposes the same functions as Standard Logger
type SpanLogger struct {
	logger  *log.Logger
	context echo.Context
}

// Writer custom writer
type Writer struct {
	io.Writer
	timeFormat string
}

func (w Writer) Write(b []byte) (n int, err error) {
	return w.Writer.Write(append([]byte(time.Now().Format(w.timeFormat)), b...))
}

//NewSpanLogger returns a new custom logger instance with TraceId and SpanId injected.
//It exposes the same functions as the standard Golang Logger, with added function of injecting span info info to the logs.
//And sending span logs to wavefront.
func NewSpanLogger(context echo.Context) *SpanLogger {
	logPrefix := context.Get("tracePrefix")
	Log := &SpanLogger{
		logger:  log.New(&Writer{os.Stdout, "2006-01-02 15:04:05 "}, fmt.Sprintf("%v", logPrefix), 0),
		context: context,
	}

	return Log
}

//Println formats using the default formats for its operands and writes to standard output.
//Spaces are always added between operands and a newline is appended.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Println(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Println(v...)

}

//Print formats using the default formats for its operands and writes to standard output.
//Spaces are added between operands when neither is a string.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Print(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Print(v...)
}

//Printf formats according to a format specifier and writes to standard output.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Printf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Printf(format, v...)
}

//Fatalln is equivalent to Println() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Fatalln(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatalln(v...)

}

//Fatal is equivalent to Print() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Fatal(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatal(v...)
}

//Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Fatalf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatalf(format, v...)
}

//Panicln is equivalent to Println() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Panicln(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panicln(v...)

}

//Panic is equivalent to Print() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Panic(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panic(v...)
}

//Panicf is equivalent to Printf() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *SpanLogger) Panicf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panicf(format, v...)
}

//Flags returns the output flags for the  logger.
func (wavefrontLogger *SpanLogger) Flags() int {
	return wavefrontLogger.logger.Flags()
}

//Prefix returns the output prefix for the logger.
func (wavefrontLogger *SpanLogger) Prefix() string {
	return wavefrontLogger.logger.Prefix()
}

//SetPrefix sets the output prefix in addition to trace Prefix for the logger.
func (wavefrontLogger *SpanLogger) SetPrefix(prefix string) {
	tracePrefix := wavefrontLogger.logger.Prefix()
	newTracePrefix := tracePrefix + prefix
	wavefrontLogger.logger.SetPrefix(newTracePrefix)
}

//SetFlags sets the output flags for the logger.
func (wavefrontLogger *SpanLogger) SetFlags(flag int) {
	wavefrontLogger.logger.SetFlags(flag)
}

//SetOutput sets the output destination for the logger.
func (wavefrontLogger *SpanLogger) SetOutput(w io.Writer) {
	wavefrontLogger.logger.SetOutput(w)
}

//Output writes the output for a logging event.
//The string s contains the text to print after the prefix specified by the flags of the Logger.
//A newline is appended if the last character of s is not already a newline.
//Calldepth is the count of the number of frames to skip when computing the file name and line number
//if Llongfile or Lshortfile is set; a value of 1 will print the details for the caller of Output.
func (wavefrontLogger *SpanLogger) Output(calldepth int, s string) error {
	return wavefrontLogger.logger.Output(calldepth, s)
}

func sendSpanLog(context echo.Context, v interface{}) {
	spanPointerInterface := context.Get("spanPointer")
	if spanPointerInterface == nil {
		return
	}

	spanPointer := spanPointerInterface.(*opentracing.Span)
	(*spanPointer).LogKV("log", v)
}
