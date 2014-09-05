// GoSrv is a thin wrapper around Go's HTTP server, to provide basic
// command-line functionality, env-specific configuration, request logging, 
// graceful shutdowns, and daemonization.
package gosrv

import (
  "fmt"
  "time"
  "os"
  "io/ioutil"
  "path/filepath"
  "strconv"
  "strings"
  "errors"
)


// Setting ForceProdEnv to true disables the -e command line argument
// and runs the app with the "prod" environment by default.
var ForceProdEnv      = false
var DefaultAddr       = ":9000"
var DefaultEnv        = "dev"
var DefaultPidFile    = ""
var DefaultConfigFile = "server.cfg"
var DefaultAppDir     = "./"
var DefaultAppName    = "server"


func stopProcessAt(pid_file string) error {
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

  // Check every 1000 ms up to 10 times (10s total)
  for i := 0; err == nil && i < 10; i++ {
    time.Sleep(1000 * time.Millisecond)
    _, err = os.Stat(pid_file)
  }

  if err == nil {
    return mkerr("Process %d is taking too long to stop.", pid) }

  return nil
}


func mkerr(msg string, obj ...interface{}) error {
  return errors.New( fmt.Sprintf(msg+"\n", obj...) )
}


func exit(status int, msg string, obj ...interface{}) {
  fmt.Printf(msg+"\n", obj...)
  os.Exit(status)
}


func daemonize(args []string) {
  procArgs := []string{}
  opt := ""

  for _, arg := range args {
    if strings.HasPrefix(arg, "-") { opt = arg }
    if opt != "-d" && opt != "-stop" && opt != "-restart" {
      procArgs = append(procArgs, arg)
    }
  }

  procName := filepath.Base(procArgs[0])
  procArgs[0] = procName
  procAttr := &os.ProcAttr{
    Dir: DefaultAppDir,
    Env: os.Environ(),
    Files: []*os.File{ os.Stdin, os.Stdout, os.Stderr },
  }

  fmt.Println("Starting daemon "+procName+"...")
  _, err := os.StartProcess(procName, procArgs, procAttr)
  if err != nil { exit(1, err.Error()) }

  exit(0, "Done!")
}


func setDefaults(args []string) {
  path, err := filepath.Abs(args[0])
  if err != nil { panic(err) }

  DefaultAppDir     = filepath.Dir(path)
  DefaultAppName    = filepath.Base(args[0])

  DefaultPidFile    = filepath.Join(DefaultAppDir, DefaultAppName + ".pid")
  DefaultConfigFile = filepath.Join(DefaultAppDir, DefaultAppName + ".cfg")
}


func init() {
  setDefaults(os.Args)
}
