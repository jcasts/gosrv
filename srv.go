
package gosrv

import (
  "fmt"
  "time"
  "os"
  "os/signal"
  "io/ioutil"
  "path/filepath"
  "strconv"
  "errors"
)


var DefaultAddr       = ":9000"
var DefaultEnv        = "dev"
var ForceProdEnv      = false
var DefaultPidFile    = "server.pid"
var DefaultConfigFile = "server.cfg"
var DefaultAppDir     = "./"
var DefaultAppName    = "server"

var srvChan = make(chan Server, 1)


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


func setDefaults() {
  path, err := filepath.Abs(os.Args[0])
  if err != nil { panic(err) }

  DefaultAppDir     = filepath.Dir(path)
  DefaultAppName    = filepath.Base(os.Args[0])
  DefaultPidFile    = filepath.Join(DefaultAppDir, DefaultAppName + ".pid")
  DefaultConfigFile = filepath.Join(DefaultAppDir, DefaultAppName + ".cfg")
}


func init() {
  setDefaults()
  handleInterrupt()
}
