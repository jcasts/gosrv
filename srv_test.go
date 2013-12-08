package gosrv

import (
  "testing"
  "path/filepath"
  "os/exec"
  "os"
  "time"
)


var testOnStopFile = "example/done.txt"
var testDaemon = "example/main"


func testCleanupServer() {
  pwd, _ := os.Getwd()
  filename := filepath.Join(pwd, testOnStopFile)
  _, err := os.Stat(filename)
  if err != nil { return }
  err = os.Remove(filename)
  if err != nil { panic( err ) }
}


func testStartDaemon() {
  pwd, _ := os.Getwd()
  filename := filepath.Join(pwd, testDaemon)

  err := exec.Command("go", "build", "-o", filename, filename+".go").Run()
  if err != nil { panic( "Could not build test server "+testDaemon ) }

  err = exec.Command(filename, "-d", "-e", "test").Run()
  if err != nil { panic( "Could not start test server on port 9000" ) }
  time.Sleep(100 * time.Millisecond)
}


func TestDefaultVars(t *testing.T) {
  testAssertEqual(t, ":9000", DefaultAddr)

  path, err := filepath.Abs(os.Args[0])
  if err != nil { t.Fatal( err ) }
  dir := filepath.Dir(path)

  testAssertEqual(t, dir, DefaultAppDir)
  testAssertEqual(t, path+".pid", DefaultPidFile)
  testAssertEqual(t, path+".cfg", DefaultConfigFile)
  testAssertEqual(t, "gosrv.test", DefaultAppName)
  testAssertEqual(t, "dev", DefaultEnv)
  testAssertEqual(t, false, ForceProdEnv)
}


func TestStopProcessAt(t *testing.T) {
  defer testCleanupServer()
  testStartDaemon()

  pwd, _ := os.Getwd()

  pidfile := filepath.Join(pwd, testDaemon+".pid")
  err := stopProcessAt(pidfile)
  if err != nil { t.Fatal( err ) }

  _, err = os.Stat(pidfile)
  if err == nil { t.Fatal( "Test server should have removed its pid file" ) }

  filename := filepath.Join(pwd, testOnStopFile)
  _, err = os.Stat(filename)
  if err != nil { t.Fatal( "Test server should have written to "+testOnStopFile ) }
}
