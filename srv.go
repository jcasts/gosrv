
package gosrv

import (
  "fmt"
  "net"
  "net/http"
  "time"
  "os"
  "os/signal"
  "io/ioutil"
  //"path/filepath"
  "strconv"
  "errors"
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


var DefaultAddr       = ":9000"
var DefaultPidFile    = os.Args[0] + ".pid"
var DefaultConfigFile = os.Args[0] + ".cfg"
var DefaultEnv        = "dev"
var ForceProdEnv      = false

var srvChan = make(chan Server, 1)


func NewServer() *Server {
  env := DefaultEnv
  if ForceProdEnv { env = "prod" }

  mux := http.NewServeMux()
  srv := http.Server{
    Addr:     DefaultAddr,
    Handler:  mux,
  }

  s := &Server{
    Server:     &srv,
    ServeMux:   mux,
    PidFile:    DefaultPidFile,
    Env:        env,
    callbacks:  make([]func()error, 0),
  }

  return s
}


func NewServerFromConfig(config_file string, env ...string) (*Server, error) {
  s := NewServer()

  if len(env) > 0 && env[0] != "" { s.Env = env[0] }

  cfg, err := ReadConfig(config_file, s.Env)
  if err != nil { return nil, err }

  pidFile, err := cfg.String("pidFile")
  if err != nil { s.PidFile = pidFile }

  readTimeout, _ := cfg.String("readTimeout")
  rt, err := time.ParseDuration(readTimeout)
  if err != nil { s.ReadTimeout = rt }

  writeTimeout, _ := cfg.String("writeTimeout")
  wt, err := time.ParseDuration(writeTimeout)
  if err != nil { s.WriteTimeout = wt }

  maxHeaderBytes, err := cfg.Int("maxHeaderBytes")
  if err != nil { s.MaxHeaderBytes = maxHeaderBytes }

  certFile, err := cfg.String("certFile")
  if err != nil { s.CertFile = certFile }

  keyFile, err := cfg.String("keyFile")
  if err != nil { s.KeyFile = keyFile }

  return s, nil
}


func NewServerFromFlag() *Server {
  f := parseFlag()

  env := ""
  if !ForceProdEnv { env = f.env }

  s, err := NewServerFromConfig(f.configFile, env)

  if err != nil {
    if f.configFile != "" && f.configFile != DefaultConfigFile {
      exit(1, err.Error()) }
    s = NewServer()
  }

  if f.pidFile != "" && f.pidFile != DefaultPidFile { s.PidFile = f.pidFile }
  if f.addr != "" && f.addr != DefaultAddr { s.Server.Addr = f.addr }

  if f.stopServer || f.restartServer {
    fmt.Println("Stopping server...")

    err := s.StopOther()
    if err != nil { exit(1, err.Error()) }

    fmt.Println("Server stopped!")
    if f.stopServer { os.Exit(0) }
  }

  return s
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


func handleInterrupt() {
  c := make(chan os.Signal)
  signal.Notify(c, os.Interrupt)

  go func() {
    for sig := range c {
      fmt.Printf("\nCaptured %v, exiting..\n", sig)

      max := len(srvChan)
      for i := 0; i < max; i++ {
        s := <-srvChan
        s.finish()
      }
      os.Exit(0)
    }
  }()
}


func StopServerAt(pid_file string) error {
  _, err := os.Stat(pid_file)
  if err != nil {
    return mkerr("Could not stop server. PID file %s does not exists.", pid_file)}

  bytes, err := ioutil.ReadFile(pid_file)
  if err != nil {
    return mkerr("Could not stop server. PID file %s is unreadable.", pid_file) }

  pid, err := strconv.Atoi(string(bytes))
  if err != nil {
    return mkerr("Could not stop server. PID file %s is invalid.", pid_file) }

  proc, err := os.FindProcess(pid)
  if err != nil {
    return mkerr("Could not stop server. PID %d is invalid.", pid) }

  err = proc.Signal(os.Interrupt)
  if err != nil {
    return mkerr("Could not stop server. PID %d was unresponsive.", pid) }

  for i := 0; err == nil && i < 20; _, err = os.FindProcess(pid) {
    time.Sleep(100 * time.Millisecond)
  }

  if err == nil {
    return mkerr("Server at PID %d is taking too long to stop.", pid)}

  return nil
}


func mkerr(msg string, obj ...interface{}) error {
  return errors.New( fmt.Sprintf(msg+"\n", obj...) )
}


func exit(status int, msg string, obj ...interface{}) {
  fmt.Printf(msg+"\n", obj...)
  os.Exit(status)
}


func init() {
  handleInterrupt()
}
