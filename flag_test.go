package gosrv

import (
  "testing"
  "os"
)


func TestParseFlag(t *testing.T) {
  args := []string{"test","-p",":7000","-pid","path/to/server.pid",
            "-c","path/to/server.cfg","-e","stage"}

  fl := parseFlag(args)

  testAssertEqual(t, "stage", fl.env)
  testAssertEqual(t, ":7000", fl.addr)
  testAssertEqual(t, "path/to/server.pid", fl.pidFile)
  testAssertEqual(t, "path/to/server.cfg", fl.configFile)

  testAssertEqual(t, false, fl.stopServer)
  testAssertEqual(t, false, fl.restartServer)
}


func TestParseFlagForceProd(t *testing.T) {
  oldForceProd := ForceProdEnv
  defer func(){ ForceProdEnv = oldForceProd }()
  ForceProdEnv = true

  args := []string{"test","-p",":7000","-pid","path/to/server.pid",
            "-c","path/to/server.cfg"}

  fl := parseFlag(args)
  testAssertEqual(t, "", fl.env)
}


func TestParseFlagDefaults(t *testing.T) {
  args := []string{"test"}

  fl := parseFlag(args)

  testAssertEqual(t, "dev", fl.env)
  testAssertEqual(t, ":9000", fl.addr)
  testAssertEqual(t, os.Args[0]+".pid", fl.pidFile)
  testAssertEqual(t, os.Args[0]+".cfg", fl.configFile)
  testAssertEqual(t, false, fl.stopServer)
  testAssertEqual(t, false, fl.restartServer)
}
