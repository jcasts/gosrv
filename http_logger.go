package gosrv

import (
  "io"
  "net"
  "net/http"
  "time"
  "strings"
  "fmt"
)

type LogValueMapper func(*Response, *http.Request)string;

var LogValueMap = map[string]LogValueMapper {
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


type HttpLogger struct {
  logFormat   string
  TimeFormat  string
  keys        []string
  Writer      io.Writer
}


func NewHttpLogger(wr io.Writer, formats ...string) *HttpLogger {
  log_format  := DefaultLogFormat
  time_format := DefaultTimeFormat

  if len(formats) > 0 { log_format = formats[0] }
  if len(formats) > 1 { time_format = formats[1] }

  l := &HttpLogger{TimeFormat: time_format, Writer: wr}
  l.SetLogFormat(log_format)
  return l
}


func (l *HttpLogger) SetLogFormat(log_format string) {
  keys := []string{}
  for k, _ := range LogValueMap {
    if strings.Contains(log_format, k) {
      keys = append(keys, k)
    }
  }
  l.logFormat = log_format
  l.keys = keys
}



func (l *HttpLogger) Log(res *Response, req *http.Request) {
  repl := []string{}

  for _, k := range l.keys {
    repl = append(repl, k, LogValueMap[k](res, req))
  }

  r := strings.NewReplacer(repl...)
  line := r.Replace(l.logFormat) + "\n"

  l.Writer.Write([]byte(line))
}



func lvRemoteAddr(res *Response, req *http.Request) string {
  host, _, err := net.SplitHostPort(req.RemoteAddr)
  if err != nil { host = req.RemoteAddr }
  return host
}


func lvDuration(res *Response, req *http.Request) string {
  duration := time.Since(res.requestTime) / time.Microsecond
  return fmt.Sprintf("%d", duration)
}


func lvProtocol(res *Response, req *http.Request) string {
  return req.Proto
}


func lvRequestTime(res *Response, req *http.Request) string {
  return res.requestTime.Format(DefaultTimeFormat)
}


func lvRequestMethod(res *Response, req *http.Request) string {
  return req.Method
}


func lvResponseBytes(res *Response, req *http.Request) string {
  b := fmt.Sprintf("%d", res.ContentLength())
  if b == "0" { b = "-" }
  return b
}


func lvRemoteUser(res *Response, req *http.Request) string {
  remoteUser := "-"
  if req.URL.User != nil && req.URL.User.Username() != "" {
    remoteUser = req.URL.User.Username()
  } else if len(req.Header["Remote-User"]) > 0 {
    remoteUser = req.Header["Remote-User"][0]
  }
  return remoteUser
}


func lvRequestPath(res *Response, req *http.Request) string {
  return req.URL.Path
}


func lvRequestUri(res *Response, req *http.Request) string {
  return req.RequestURI
}


func lvRequestFirstLine(res *Response, req *http.Request) string {
  return lvRequestMethod(res, req) + " " +
          lvRequestUri(res, req) + " " +
          lvProtocol(res, req)
}


func lvResponseStatus(res *Response, req *http.Request) string {
  return fmt.Sprintf("%d", res.Status)
}


func lvReferer(res *Response, req *http.Request) string {
  return req.Referer()
}


func lvUserAgent(res *Response, req *http.Request) string {
  return req.UserAgent()
}
