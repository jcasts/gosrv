package gosrv

import (
  "io"
  "net"
  "net/http"
  "time"
  "strings"
  "fmt"
)

// Interface for value fetching functions. Time represents when the request
// was received. The given ResponseWriter should always be a gosrv.Response
// when used in a gosrv.Mux.
type LogValueFunc func(time.Time, http.ResponseWriter, *http.Request)string;

// Map of logFormat keywords to functions. More may be added at need.
var LogValueMap = map[string]LogValueFunc {
  "$RemoteAddr": lvRemoteAddr,
  "$Protocol": lvProtocol,
  "$RequestTime": lvDuration,
  "$Time": lvRequestTime,
  "$RequestMethod": lvRequestMethod,
  "$BodyBytes": lvResponseBytes,
  "$RemoteUser": lvRemoteUser,
  "$RequestUri": lvRequestUri,
  "$RequestPath": lvRequestPath,
  "$Request": lvRequestFirstLine,
  "$Status": lvResponseStatus,
  "$HttpReferer": lvReferer,
  "$HttpUserAgent": lvUserAgent,
}

var DefaultLogFormat =
  "$RemoteAddr - $RemoteUser $Time \"$Request\" $Status $BodyBytes \"$HttpReferer\" \"$HttpUserAgent\""
var DefaultTimeFormat = "[02/Jan/2006:15:04:05 -0700]"


// The logger interface for gosrv.Mux.
type HttpLogger interface {
  io.Writer
  SetLogFormat(format string)
  SetTimeFormat(time_format string)
  SetWriter(wr io.Writer)
  Log(t time.Time, wr http.ResponseWriter, req *http.Request)
}


type httpLogger struct {
  logFormat   string
  timeFormat  string
  keys        []string
  writer      io.Writer
}


func NewHttpLogger(wr io.Writer, formats ...string) HttpLogger {
  log_format  := DefaultLogFormat
  time_format := DefaultTimeFormat

  if len(formats) > 0 { log_format = formats[0] }
  if len(formats) > 1 { time_format = formats[1] }

  l := &httpLogger{timeFormat: time_format, writer: wr}
  l.SetLogFormat(log_format)
  return l
}


func (l *httpLogger) SetLogFormat(log_format string) {
  keys := []string{}
  for k, _ := range LogValueMap {
    if strings.Contains(log_format, k) {
      keys = append(keys, k)
    }
  }
  l.logFormat = log_format
  l.keys = keys
}


func (l *httpLogger) SetTimeFormat(time_format string) {
  l.timeFormat = time_format
}


func (l *httpLogger) SetWriter(wr io.Writer) {
  l.writer = wr
}


func (l *httpLogger) Write(bytes []byte) (int, error) {
  return l.writer.Write(bytes)
}


func (l *httpLogger) Log(t time.Time, wr http.ResponseWriter, req *http.Request) {
  repl := []string{}

  for _, k := range l.keys {
    repl = append(repl, k, LogValueMap[k](t, wr, req))
  }

  r := strings.NewReplacer(repl...)
  line := r.Replace(l.logFormat) + "\n"

  l.Write([]byte(line))
}



func lvRemoteAddr(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  host, _, err := net.SplitHostPort(req.RemoteAddr)
  if err != nil { host = req.RemoteAddr }
  return host
}


func lvDuration(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  duration := time.Since(t) / time.Microsecond
  return fmt.Sprintf("%d", duration)
}


func lvProtocol(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.Proto
}


func lvRequestTime(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return t.Format(DefaultTimeFormat)
}


func lvRequestMethod(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.Method
}


func lvResponseBytes(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  b := "-"
  if res, ok := wr.(*Response); ok {
    b = fmt.Sprintf("%d", res.ContentLength())
    if b == "0" { b = "-" }
  }
  return b
}


func lvRemoteUser(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  remoteUser := "-"
  if req.URL.User != nil && req.URL.User.Username() != "" {
    remoteUser = req.URL.User.Username()
  } else if len(req.Header["Remote-User"]) > 0 {
    remoteUser = req.Header["Remote-User"][0]
  }
  return remoteUser
}


func lvRequestPath(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.URL.Path
}


func lvRequestUri(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.RequestURI
}


func lvRequestFirstLine(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return lvRequestMethod(t, wr, req) + " " +
          lvRequestUri(t, wr, req) + " " +
          lvProtocol(t, wr, req)
}


func lvResponseStatus(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  if res, ok := wr.(*Response); ok {
    return fmt.Sprintf("%d", res.Status) }
  return "-"
}


func lvReferer(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.Referer()
}


func lvUserAgent(t time.Time, wr http.ResponseWriter, req *http.Request) string {
  return req.UserAgent()
}
