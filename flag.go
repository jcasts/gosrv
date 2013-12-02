package gosrv

include (
  "fmt"
  "flag"
  "os"
)


type parsedFlag struct {
  stopServer    bool
  restartServer bool
  env           string
  addr          string
  configFile    string
  pidFile       string
}


func parseFlag() *parsedFlag {
  f := &parsedFlag{}
  setFlag(f)
  flag.Usage = flagUsage
  flag.Parse()
  return f
}


func setFlag(f *parsedFlag) {
  flag.StringVar(f.addr, "p", DefaultAddr, "\tServer address")
  flag.StringVar(f.pidFile, "pid", DefaultPidFile, "\tServer PID File")
  flag.StringVar(f.configFile, "c", DefaultConfigFile, "\tConfig file")

  if !ForceProdEnv {
    flag.StringVar(f.env, "e", DefaultEnv, "\tEnvironment to run server in") }

  flag.BoolVar(f.stopServer, "stop", false, "\tStop running server and exit")
  flag.BoolVar(f.restartServer, "restart", false, "\tStop running server and boot")
}


func flagUsage() {
  fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
  flag.PrintDefaults()
  os.Exit(2)
}

