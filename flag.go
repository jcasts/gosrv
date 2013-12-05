package gosrv

import (
  "fmt"
  "flag"
  "os"
)


type parsedFlag struct {
  daemonizeServer bool
  stopServer      bool
  restartServer   bool
  env             string
  addr            string
  configFile      string
  pidFile         string
  flagSet         *flag.FlagSet
}


func parseFlag(args []string) *parsedFlag {
  f := parsedFlag{}
  f.setFlag(DefaultAppName)
  f.flagSet.Parse(args[1:])
  return &f
}


func (f *parsedFlag) setFlag(name string) {
  flagset := flag.NewFlagSet(name, flag.ExitOnError)
  flagset.StringVar(&f.addr, "a", DefaultAddr, "\tServer address")
  flagset.StringVar(&f.pidFile, "pid", DefaultPidFile, "\tServer PID File")
  flagset.StringVar(&f.configFile, "c", DefaultConfigFile, "\tConfig file")

  if !ForceProdEnv {
    flagset.StringVar(&f.env, "e", DefaultEnv, "\tEnvironment to run server in") }

  flagset.BoolVar(&f.daemonizeServer, "d", false, "\tRun server as daemon")
  flagset.BoolVar(&f.stopServer, "stop", false, "\tStop running server and exit")
  flagset.BoolVar(&f.restartServer, "restart", false, "\tStop running server and boot daemon")

  flagset.Usage = func() {
    fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", name)
    flagset.PrintDefaults()
    os.Exit(2)
  }

  f.flagSet = flagset
}
