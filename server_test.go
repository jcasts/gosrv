package gosrv

import (
  "testing"
  "time"
)


func TestNewServer(t *testing.T) {
  s := NewServer()
  testAssertEqual(t, "dev", s.Env)
  testAssertEqual(t, ":9000", s.Addr)
}


func TestNewServerForceProd(t *testing.T) {
  oldForceProd := ForceProdEnv
  defer func(){ ForceProdEnv = oldForceProd }()
  ForceProdEnv = true

  s := NewServer()
  testAssertEqual(t, "prod", s.Env)
  testAssertEqual(t, ":9000", s.Addr)
}


func TestNewServerFromConfig(t *testing.T) {
  s, err := NewServerFromConfig("testdata/server.cfg")
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


func TestNewServerFromConfigEnv(t *testing.T) {
  s, err := NewServerFromConfig("testdata/server.cfg", "prod")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "prod", s.Env)
  testAssertEqual(t, "path/to/server.pid", s.PidFile)
  testAssertEqual(t, 1 * time.Second, s.ReadTimeout)
  testAssertEqual(t, 1 * time.Second, s.WriteTimeout)
  testAssertEqual(t, 1024, s.MaxHeaderBytes)
  testAssertEqual(t, "path/to/server.cert", s.CertFile)
  testAssertEqual(t, "path/to/server.key", s.KeyFile)

  val, err := s.Config.String("customConfig")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "customValue", val)
}


func TestNewServerFromFlag(t *testing.T) {
  s, err := NewServerFromFlag(
              "test","-a",":7000","-pid","path/to/server.pid","-e","stage")
  if err != nil { t.Fatal( err ) }

  testAssertEqual(t, "stage", s.Env)
  testAssertEqual(t, "stage", s.Config.Env)
  testAssertEqual(t, ":7000", s.Addr)
  testAssertEqual(t, "path/to/server.pid", s.PidFile)
}


func TestNewServerFromFlagAndConfig(t *testing.T) {
  s, err := NewServerFromFlag(
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