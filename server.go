package gosrv

import (
  "fmt"
  "net"
  "net/http"
  "os"
  "io/ioutil"
  "time"
)


// The gosrv server is a wrapper around Go's http server and http handler.
// It handles Interrupt signals with callbacks before shutdown, and can be
// constructed via a config file, command line options, or both.
type Server struct {
  *http.Server
  *Mux
  Config      *Config
  PidFile     string
  Env         string
  CertFile    string
  KeyFile     string
  callbacks   []func()error
}


// Creates a new Server instance with an optional environment name.
// The default environment is "dev".
func New(env ...string) *Server {
  s := &Server{PidFile: DefaultPidFile}

  if len(env) > 0 && env[0] != "" {
    s.Env = env[0]
  } else if ForceProdEnv {
    s.Env = "prod"
  } else {
    s.Env = DefaultEnv
  }

  mux := NewMux()

  s.Server = &http.Server{
    Addr:     DefaultAddr,
    Handler:  mux,
  }

  s.Mux  = mux
  s.Config    = NewConfig(s.Env)
  s.callbacks = make([]func()error, 0)

  return s
}


// Create a new Server from a config file with an optional environment name.
// The default environment is "dev".
//
// Valid configuration keys are:
//  * addr            The address to boot the server on (default ":9000")
//  * pidFile         Location to store PID in (default "<bin-name>.pid")
//  * readTimeout     Server read timeout (default to net/http default)
//  * writeTimeout    Server write timeout (default to net/http default)
//  * maxHeaderBytes  Max header bytes allowed (default to net/http default)
//  * logFormat       Log format to write in (default to DefaultLogFormat)
//  * timeFormat      Time format for logs (default to DefaultTimeFormat)
//  * certFile        TLS cert file (default none)
//  * keyFile         TLS key file (default none)
func NewFromConfig(config_file string, env ...string) (*Server, error) {
  s := New()

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

  addr, err := cfg.String("addr")
  if err == nil { s.Addr = addr }

  logFormat, err := cfg.String("logFormat")
  if err == nil { s.Logger.SetLogFormat(logFormat) }

  timeFormat, err := cfg.String("timeFormat")
  if err == nil { s.Logger.SetTimeFormat(timeFormat) }

  certFile, err := cfg.String("certFile")
  if err == nil { s.CertFile = certFile }

  keyFile, err := cfg.String("keyFile")
  if err == nil { s.KeyFile = keyFile }

  return s, nil
}


// Reads command line arguments to create a new Server instance. Uses a
// config file if provided to the -c option. Command line arguments
// override config values.
func NewFromFlag(args ...string) (*Server, error) {
  var s *Server
  var err error

  if len(args) == 0 { args = os.Args }
  f := parseFlag(args)

  env := ""
  if !ForceProdEnv { env = f.env }

  if f.configFile != "" {
    s, err = NewFromConfig(f.configFile, env)
    if err != nil && f.configFile != DefaultConfigFile { return s, err }
  }

  if s == nil { s = New(env) }

  if f.pidFile != "" && f.pidFile != DefaultPidFile { s.PidFile = f.pidFile }
  if f.addr != "" && f.addr != DefaultAddr { s.Addr = f.addr }

  if f.stopServer || f.restartServer {
    fmt.Println("Stopping server...")

    err := s.StopOther()
    if err != nil { return s, err }

    fmt.Println("Server stopped!")
    if f.stopServer { os.Exit(0) }
  }

  if f.restartServer || f.daemonizeServer {
    daemonize(args)
    os.Exit(0)
  }

  return s, nil
}


// Starts the server and listens on the given server.Addr.
func (s *Server) ListenAndServe() error {
  err := s.prepare()
  if err != nil { return err }

  if s.CertFile != "" && s.KeyFile != "" {
    return s.Server.ListenAndServeTLS(s.CertFile, s.KeyFile)
  } else {
    return s.Server.ListenAndServe()
  }
}


// Starts the server with the given TLS files and listens on server.Addr.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
  err := s.prepare()
  if err != nil { return err }
  return s.Server.ListenAndServeTLS(certFile, keyFile)
}


// Starts the server for the given listener.
func (s *Server) Serve(l net.Listener) error {
  err := s.prepare()
  if err != nil { return err }
  return s.Server.Serve(l)
}


// Writes the server's pidfile. Typically called at server Listen time.
func (s Server) WritePidFile() error {
  _, err := os.Stat(s.PidFile)

  if err == nil {
    return mkerr("PID file %s already exists. Please delete it and try again.", s.PidFile) }

  content := []byte( fmt.Sprintf("%d", os.Getpid()) )
  err = ioutil.WriteFile(s.PidFile, content, 0666)
  return err
}


// Removes the server's pidfile. The pidfile is automatically deleted when
// an interrupt signal is received.
func (s Server) DeletePidFile() error {
  _, err := os.Stat(s.PidFile)
  if err != nil { return nil }
  return os.Remove(s.PidFile)
}


// Stop the server running at server.PidFile.
func (s *Server) StopOther() error {
  return stopProcessAt(s.PidFile)
}


// Add a callback for when the server shuts down.
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
