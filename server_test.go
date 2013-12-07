package gosrv

import (
  "testing"
  "time"
)


func TestNew(t *testing.T) {
  s := New()
  testAssertEqual(t, "dev", s.Env)
  testAssertEqual(t, "", s.Addr)
}


func TestNewForceProd(t *testing.T) {
  oldForceProd := ForceProdEnv
  defer func(){ ForceProdEnv = oldForceProd }()
  ForceProdEnv = true

  s := New()
  testAssertEqual(t, "prod", s.Env)
  testAssertEqual(t, "", s.Addr)
}


func TestNewFromConfig(t *testing.T) {
  s, err := NewFromConfig("testdata/server.cfg")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "dev", s.Env)
  testAssertEqual(t, DefaultPidFile, s.PidFile)
  testAssertEqual(t, 1 * time.Second, s.ReadTimeout)
  testAssertEqual(t, 500 * time.Millisecond, s.WriteTimeout)
  testAssertEqual(t, 1024, s.MaxHeaderBytes)
  testAssertEqual(t, "", s.CertFile)
  testAssertEqual(t, "", s.KeyFile)

  val, err := s.Config.String("customConfig")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "customValue", val)
}


func TestNewFromConfigEnv(t *testing.T) {
  s, err := NewFromConfig("testdata/server.cfg", "prod")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "prod", s.Env)
  testAssertEqual(t, "path/to/server.pid", s.PidFile)
  testAssertEqual(t, 1 * time.Second, s.ReadTimeout)
  testAssertEqual(t, 1 * time.Second, s.WriteTimeout)
  testAssertEqual(t, 1024, s.MaxHeaderBytes)
  testAssertEqual(t, "path/to/server.cert", s.CertFile)
  testAssertEqual(t, "path/to/server.key", s.KeyFile)

  l, _ := s.Logger.(*httpLogger)
  testAssertEqual(t, "(02/01/2006 15:04:05)", l.timeFormat)

  logFmt := "$RemoteAddr - $RemoteUser $Time \"$Request\" $Status $BodyBytes"
  testAssertEqual(t, logFmt, l.logFormat)

  val, err := s.Config.String("customConfig")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "customValue", val)
}


func TestNewFromFlag(t *testing.T) {
  s, err := NewFromFlag(
              "test","-a",":7000","-pid","path/to/server.pid","-e","stage")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "stage", s.Env)
  testAssertEqual(t, "stage", s.Config.Env)
  testAssertEqual(t, ":7000", s.Addr)
  testAssertEqual(t, "path/to/server.pid", s.PidFile)
}


func TestNewFromFlagAndConfig(t *testing.T) {
  s, err := NewFromFlag(
              "test","-a",":7000","-c","testdata/server.cfg","-e","prod")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "prod", s.Env)
  testAssertEqual(t, "prod", s.Config.Env)
  testAssertEqual(t, ":7000", s.Addr)

  testAssertEqual(t, 1024, s.MaxHeaderBytes)
  testAssertEqual(t, 1 * time.Second, s.WriteTimeout)

  val, err := s.Config.String("customConfig")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "customValue", val)
}
