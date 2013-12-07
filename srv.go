// GoSrv is a thin wrapper around GoLang's HTTP server, to provide basic
// command-line functionality, env-specific configuration, request logging,
// and graceful shutdowns.
package gosrv

import (
  "fmt"
  "time"
  "os"
  "os/signal"
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
var DefaultPidFile    = "server.pid"
var DefaultConfigFile = "server.cfg"
var DefaultAppDir     = "./"
var DefaultAppName    = "server"

var srvChan = make(chan *Server, 1)
var sigChan = make(chan os.Signal)


func handleInterrupt() {
  signal.Notify(sigChan, os.Interrupt)

  go func() {
    for sig := range sigChan {
      fmt.Printf("\nCaptured %v, exiting..\n", sig)

      max := len(srvChan)
      for i := 0; i < max; i++ {
        s := <-srvChan
        s.Stop()
      }
    }
  }()
}


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

  for i := 0; err == nil && i < 20; i++ {
    time.Sleep(100 * time.Millisecond)
    _, err = os.Stat(pid_file)
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
  handleInterrupt()
}
