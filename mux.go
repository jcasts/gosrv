package gosrv

import (
  "net/http"
  "time"
  "os"
)


type Mux struct {
  *http.ServeMux
  Logger *HttpLogger
}


func NewMux() *Mux {
  return &Mux{http.NewServeMux(), NewHttpLogger(os.Stdout)}
}


func (m *Mux) ServeHTTP(wr http.ResponseWriter, req *http.Request) {
  res := &Response{wr, 200, 0, time.Now()}
  m.ServeMux.ServeHTTP(res, req)
  m.Logger.Log(res, req)
}
