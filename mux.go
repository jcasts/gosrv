package gosrv

import (
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
    "02/01/2006 15:04:05", "%s %s [%s] \"%s %s\" %s %d %d %0.4f\n"}
}


func (m *Mux) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
  stime := time.Now()

  res := &Response{wr, 200, 0}
  m.ServeMux.ServeHTTP(res, req)
  m.logRequest(res, req, stime)
}


func (m Mux) logRequest(res *Response, req *http.Request, stime time.Time) {
  duration := float64(time.Since(stime)) / 1000000

  var remoteUser string
  if len(req.Header["Remote-User"]) > 0 {
    remoteUser = req.Header["Remote-User"][0] }

  if remoteUser == "" { remoteUser = "-" }

  timeFormat := m.TimeFormat
  time_str := stime.Format(timeFormat)

  clen := res.ContentLength()

  logFormat := m.LogFormat
  fmt.Printf(logFormat, req.RemoteAddr, remoteUser, time_str,
    req.Method, req.RequestURI, req.Proto, res.Status, clen, duration)
}
