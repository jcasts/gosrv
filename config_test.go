package gosrv

import (
  "testing"
)


func TestReadConfig(t *testing.T) {
  c, err := ReadConfig("testdata/server.cfg", "dev")
  if err != nil { t.Fatal( err ) }

  val, err := c.String("customConfig")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "customValue", val)

  val, err = c.String("writeTimeout")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "500ms", val)

  val, err = c.String("readTimeout")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "1s", val)

  val, err = c.String("certFile")
  if err == nil { t.Fatal( "Value certFile should be empty but was "+val ) }
}


func TestConfigEnv(t *testing.T) {
  c, err := ReadConfig("testdata/server.cfg", "dev")
  if err != nil { t.Fatal( err ) }

  val, err := c.String("certFile")
  if err == nil { t.Fatal( "Value certFile should be empty but was "+val ) }

  c.Env = "prod"

  val, err = c.String("certFile")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "path/to/server.cert", val)

  val, err = c.String("writeTimeout")
  if err != nil { t.Fatal( err ) }
  testAssertEqual(t, "1s", val)
}
