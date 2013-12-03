
package gosrv

import (
  "fmt"
  "net"
  "net/http"
  "os"
  "io/ioutil"
  "time"
)


type Server struct {
  *http.Server
  *http.ServeMux
  Config      *Config
  PidFile     string
  Env         string
  CertFile    string
  KeyFile     string
  callbacks   []func()error
}


func NewServer(env ...string) *Server {
  s := &Server{PidFile: DefaultPidFile}

  if len(env) > 0 && env[0] != "" {
    s.Env = env[0]
  } else if ForceProdEnv {
    s.Env = "prod"
  } else {
    s.Env = DefaultEnv
  }

  mux := http.NewServeMux()

  s.Server = &http.Server{
    Addr:     DefaultAddr,
    Handler:  mux,
  }

  s.ServeMux  = mux
  s.PidFile   = DefaultPidFile
  s.Config    = NewConfig(s.Env)
  s.callbacks = make([]func()error, 0)

  return s
}


func NewServerFromConfig(config_file string, env ...string) (*Server, error) {
  s := NewServer()

  if len(env) > 0 && env[0] != "" { s.Env = env[0] }

  cfg, err := ReadConfig(config_file, s.Env)
  if err != nil { return nil, err }
  s.Config = cfg

  pidFile, err := cfg.String("pidFile")
  if err == nil && pidFile != "" { s.PidFile = pidFile }

  readTimeout, _ := cfg.String("readTimeout")
  rt, err := time.ParseDuration(readTimeout)
  if err == nil { s.ReadTimeout = rt }

  writeTimeout, _ := cfg.String("writeTimeout")
  wt, err := time.ParseDuration(writeTimeout)
  if err == nil { s.WriteTimeout = wt }

  maxHeaderBytes, err := cfg.Int("maxHeaderBytes")
  if err == nil { s.MaxHeaderBytes = maxHeaderBytes }

  addr, _ := cfg.String("addr")
  if err == nil { s.Addr = addr }

  certFile, err := cfg.String("certFile")
  if err == nil { s.CertFile = certFile }

  keyFile, err := cfg.String("keyFile")
  if err == nil { s.KeyFile = keyFile }

  return s, nil
}


func NewServerFromFlag(args ...string) (*Server, error) {
  var s *Server
  var err error

  if len(args) == 0 { args = os.Args }
  f := parseFlag(args)

  env := ""
  if !ForceProdEnv { env = f.env }

  if f.configFile != "" {
    s, err = NewServerFromConfig(f.configFile, env)
    if err != nil && f.configFile != DefaultConfigFile { return s, err }
  }

  if s == nil { s = NewServer(env) }

  if f.pidFile != "" && f.pidFile != DefaultPidFile { s.PidFile = f.pidFile }
  if f.addr != "" && f.addr != DefaultAddr { s.Addr = f.addr }

  if f.stopServer || f.restartServer {
    fmt.Println("Stopping server...")

    err := s.StopOther()
    if err != nil { return s, err }

    fmt.Println("Server stopped!")
    if f.stopServer { os.Exit(0) }
  }

  return s, nil
}


func (s *Server) ListenAndServe() error {
  err := s.prepare()
  if err != nil { return err }

  if s.CertFile != "" && s.KeyFile != "" {
    return s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile)
  } else {
    return s.Server.ListenAndServe()
  }
}


func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
  err := s.prepare()
  if err != nil { return err }
  return s.Server.ListenAndServeTLS(certFile, keyFile)
}


func (s *Server) Serve(l net.Listener) error {
  err := s.prepare()
  if err != nil { return err }
  return s.Server.Serve(l)
}


func (s Server) WritePidFile() error {
  _, err := os.Stat(s.PidFile)

  if err == nil {
    return mkerr("PID file %s already exists. Please delete it and try again.", s.PidFile) }

  content := []byte( fmt.Sprintf("%d", os.Getpid()) )
  err = ioutil.WriteFile(s.PidFile, content, 0666)
  return err
}


func (s Server) DeletePidFile() error {
  _, err := os.Stat(s.PidFile)
  if err != nil { return nil }
  return os.Remove(s.PidFile)
}


func (s *Server) StopOther() error {
  return StopServerAt(s.PidFile)
}


func (s *Server) OnStop(fn func()error) {
  if len(s.callbacks) == 0 {
    s.callbacks = []func()error{} }
  s.callbacks = append(s.callbacks, fn)
}


func (s *Server) prepare() error {
  s.callbacks = append(s.callbacks, s.DeletePidFile)
  err := s.WritePidFile()
  if err != nil { return err }

  srvChan <- *s
  return nil
}


func (s *Server) finish() {
  for _, fn := range s.callbacks {
    fn()
  }
}
