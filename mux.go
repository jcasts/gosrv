package gosrv

import (
  "net/http"
  "time"
  "os"
  "sync"
)


// The Mux struct handles HTTP logging and graceful server shutdown.
type Mux struct {
  *http.ServeMux
  Logger  HttpLogger
  conns   *sync.WaitGroup
}


func NewMux() *Mux {
  return &Mux{http.NewServeMux(), NewHttpLogger(os.Stdout), &sync.WaitGroup{}}
}


func (m *Mux) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
  res := &Response{wr, 200, 0}
  stime := time.Now()
  m.ServeMux.ServeHTTP(res, req)
  m.Logger.Log(stime, res, req)
  if m.conns != nil { m.conns.Done() }
}
