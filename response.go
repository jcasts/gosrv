package gosrv

import (
  "net/http"
  "strconv"
)


type Response struct {
  http.ResponseWriter
  Status int
  written int
}


func (r Response) ContentLength() int {
  clen_str := r.Header().Get("Content-Length")
  if clen_str != "" {
    clen, err := strconv.Atoi(clen_str)
    if err == nil { return clen }
  }
  return r.written
}


func (r *Response) Write(body []byte) (int, error) {
  num, err := r.ResponseWriter.Write(body)
  r.written += num
  return num, err
}


func (r *Response) WriteHeader(status int) {
  r.Status = status
  r.ResponseWriter.WriteHeader(status)
}
