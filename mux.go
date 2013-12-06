package gosrv

import (
  "net"
  "net/http"
  "time"
  "fmt"
)


type Mux struct {
  *http.ServeMux
  TimeFormat  string
  LogFormat   string
}


func NewMux() *Mux {
  return &Mux{http.NewServeMux(),
    "02/Jan/2006:15:04:05 -0700",
    "%s - %s [%s] \"%s %s %s\" %d %d %0.4f \"%s\" \"%s\"\n"}
}


func (m *Mux) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
  stime := time.Now()

  res := &Response{wr, 200, 0}
  m.ServeMux.ServeHTTP(res, req)
  m.logRequest(res, req, stime)
}


func (m Mux) logRequest(res *Response, req *http.Request, stime time.Time) {
  duration := float64(time.Since(stime)) / 1000000

  remoteUser := "-"
  if req.URL.User != nil && req.URL.User.Username() != "" {
    remoteUser = req.URL.User.Username()
  } else if len(req.Header["Remote-User"]) > 0 {
    remoteUser = req.Header["Remote-User"][0]
  }

  host, _, err := net.SplitHostPort(req.RemoteAddr)
  if err != nil { host = req.RemoteAddr }

  timeFormat := m.TimeFormat
  time_str := stime.Format(timeFormat)

  clen := res.ContentLength()

  logFormat := m.LogFormat
  fmt.Printf(logFormat, host, remoteUser, time_str,
    req.Method, req.RequestURI, req.Proto, res.Status, clen, duration,
    req.Referer(), req.UserAgent())
}
