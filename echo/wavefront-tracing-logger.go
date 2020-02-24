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

//WfLogger custom implementation of Go standard Logger
//Exposes the same functions as Standard Logger
type WfLogger struct {
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

//NewWfLogger returns a new custom logger instance with TraceId and SpanId injected.
//It exposes the same functions as the standard Golang Logger, with added function of injecting trace info in logs.
//And sending span logs to wavefront.
func NewWfLogger(context echo.Context) *WfLogger {
	logPrefix := context.Get("tracePrefix")
	Log := &WfLogger{
		logger:  log.New(&Writer{os.Stdout, "2006-01-02 15:04:05 "}, fmt.Sprintf("%v", logPrefix), 0),
		context: context,
	}

	return Log
}

//Println formats using the default formats for its operands and writes to standard output.
//Spaces are always added between operands and a newline is appended.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Println(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Println(v...)

}

//Print formats using the default formats for its operands and writes to standard output.
//Spaces are added between operands when neither is a string.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Print(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Print(v...)
}

//Printf formats according to a format specifier and writes to standard output.
//It returns the number of bytes written and any write error encountered.
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Printf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Printf(format, v...)
}

//Fatalln is equivalent to Println() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Fatalln(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatalln(v...)

}

//Fatal is equivalent to Print() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Fatal(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatal(v...)
}

//Fatalf is equivalent to Printf() followed by a call to os.Exit(1).
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Fatalf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Fatalf(format, v...)
}

//Panicln is equivalent to Println() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Panicln(v ...interface{}) {
	logObject := fmt.Sprintln(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panicln(v...)

}

//Panic is equivalent to Print() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Panic(v ...interface{}) {
	logObject := fmt.Sprint(v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panic(v...)
}

//Panicf is equivalent to Printf() followed by a call to panic().
//Sends the log object to wavefront.
func (wavefrontLogger *WfLogger) Panicf(format string, v ...interface{}) {
	logObject := fmt.Sprintf(format, v...)
	sendSpanLog(wavefrontLogger.context, logObject)
	wavefrontLogger.logger.Panicf(format, v...)
}

//Flags returns the output flags for the  logger.
func (wavefrontLogger *WfLogger) Flags() int {
	return wavefrontLogger.logger.Flags()
}

//Prefix returns the output prefix for the logger.
func (wavefrontLogger *WfLogger) Prefix() string {
	return wavefrontLogger.logger.Prefix()
}

//SetPrefix sets the output prefix in addition to trace Prefix for the logger.
func (wavefrontLogger *WfLogger) SetPrefix(prefix string) {
	tracePrefix := wavefrontLogger.logger.Prefix()
	newTracePrefix := tracePrefix + prefix
	wavefrontLogger.logger.SetPrefix(newTracePrefix)
}

//SetFlags sets the output flags for the logger.
func (wavefrontLogger *WfLogger) SetFlags(flag int) {
	wavefrontLogger.logger.SetFlags(flag)
}

//SetOutput sets the output destination for the logger.
func (wavefrontLogger *WfLogger) SetOutput(w io.Writer) {
	wavefrontLogger.logger.SetOutput(w)
}

//Output writes the output for a logging event.
//The string s contains the text to print after the prefix specified by the flags of the Logger.
//A newline is appended if the last character of s is not already a newline.
//Calldepth is the count of the number of frames to skip when computing the file name and line number
//if Llongfile or Lshortfile is set; a value of 1 will print the details for the caller of Output.
func (wavefrontLogger *WfLogger) Output(calldepth int, s string) error {
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
