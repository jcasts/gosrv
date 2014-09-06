package gosrv

import (
  "net"
)

type Listener struct {
  net.Listener
  server *Server
}

func (l Listener) Accept() (c net.Conn, err error) {
  c, err = l.Listener.Accept()

  if err == nil && l.server.conns != nil {
    l.server.conns.Add(1) }

  return c, err
}