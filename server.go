package gosrv

import (
  "fmt"
  "net"
  "net/http"
  "os"
  "io/ioutil"
  "time"
  "crypto/tls"
  "sync"
  "os/signal"
)


// The gosrv server is a wrapper around Go's http server and http handler.
// It supports shutdown via Interrupt signals, and can be
// constructed via a config file, command line options, or both.
type Server struct {
  *http.Server
  *Mux
  Config      *Config
  PidFile     string
  Env         string
  CertFile    string
  KeyFile     string
  listener    net.Listener
  stopped     bool
  rwlock      sync.RWMutex
  sigchan     chan os.Signal
}


// Creates a new Server instance with an optional environment name.
// The default environment is "dev".
func New(env ...string) *Server {
  s := &Server{PidFile: DefaultPidFile,
    rwlock: sync.RWMutex{}, sigchan: make(chan os.Signal)}

  if len(env) > 0 && env[0] != "" {
    s.Env = env[0]
  } else if ForceProdEnv {
    s.Env = "prod"
  } else {
    s.Env = DefaultEnv
  }

  mux := NewMux()
  s.Server = &http.Server{Handler: mux}
  s.Mux    = mux
  s.Config = NewConfig(s.Env)

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
//  * logFile         File to write request logs to (default stdout)
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

  logFile, err := cfg.String("logFile")
  if err == nil {
    f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
    if err != nil { return s, err }
    s.Logger.SetWriter(f)
  }

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
  if s.CertFile != "" && s.KeyFile != "" {
    return s.ListenAndServeTLS(s.CertFile, s.KeyFile)

  } else {
    if s.Addr == "" { s.Addr = DefaultAddr }

    l, e := net.Listen("tcp", s.Addr)
    if e != nil { return e }

    s.Logger.Printf("Server %s listening...\n", s.Addr)

    return s.Serve(l)
  }
}


// Starts the server with the given TLS files and listens on server.Addr.
func (s *Server) ListenAndServeTLS(certFile, keyFile string) error {
  if s.Addr == "" { s.Addr = ":https" }

  config := &tls.Config{}
  if s.TLSConfig != nil { *config = *s.TLSConfig }
  if config.NextProtos == nil {
    config.NextProtos = []string{"http/1.1"}
  }

  var err error
  config.Certificates = make([]tls.Certificate, 1)
  config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
  if err != nil { return err }

  conn, err := net.Listen("tcp", s.Addr)
  if err != nil { return err }

  s.Logger.Printf("Server %s listening...\n", s.Addr)

  tlsListener := tls.NewListener(conn, config)
  return s.Serve(tlsListener)
}


// Starts the server for the given listener.
func (s *Server) Serve(l net.Listener) error {
  s.Stop()

  err := s.prepare()

  if err == nil {
    s.listener = l
    err = s.Server.Serve(l)
  }

  if s.stopped { err = nil }
  s.stopped = false

  if err != nil { s.Logger.Printf(err.Error() + "\n") }

  return err
}


// Writes the server's pidfile. Typically called at server Listen time.
func (s Server) WritePidFile() error {
  if s.PidFile == "" { return nil }

  pidStr := fmt.Sprintf("%d", os.Getpid())

  _, err := os.Stat(s.PidFile)

  if err == nil {
    bytes, err := ioutil.ReadFile(s.PidFile)
    filePid := string(bytes)
    if err == nil && filePid == pidStr {
      return nil
    } else if filePid != "" {
      return mkerr("PID file %s already exists. Please delete it and try again.", s.PidFile)
    }
  }

  err = ioutil.WriteFile(s.PidFile, []byte(pidStr), 0666)
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


// Stop the server and gracefully shutdown connections.
func (s *Server) Stop() error {
  s.rwlock.RLock()
  if s.listener == nil {
    s.rwlock.RUnlock()
    return mkerr("Listener non-existant")
  }
  s.rwlock.RUnlock()

  s.rwlock.Lock()
  s.stopped = true
  s.listener.Close()
  s.listener = nil
  s.rwlock.Unlock()

  s.Logger.Printf("Server %s stopping...\n", s.Addr)
  return s.finish()
}


// Returns true if the server is running.
func (s *Server) Running() bool {
  s.rwlock.RLock()
  r := s.listener != nil && !s.stopped
  s.rwlock.RUnlock()

  return r
}


func (s *Server) prepare() error {
  err := s.WritePidFile()
  if err != nil { return err }

  s.conns.Add(1)

  signal.Notify(s.sigchan, os.Interrupt)
  go func() {
    <-s.sigchan // block until signal is received
    s.Stop()
  }()

  return nil
}


func (s *Server) finish() error {
  s.conns.Done()
  signal.Stop(s.sigchan)
  close(s.sigchan)
  return s.DeletePidFile()
}
