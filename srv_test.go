package gosrv

import (
  "testing"
  "path/filepath"
  "os"
)


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
