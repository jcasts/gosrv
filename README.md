[![Go Walker](http://gowalker.org/api/v1/badge)](http://gowalker.org/github.com/jcasts/gosrv)
[![githalytics.com alpha](https://cruel-carlota.pagodabox.com/219643edaed8bef045626eb7dd5418cf "githalytics.com")](http://githalytics.com/jcasts/gosrv)
# GoSrv

GoSrv is a thin wrapper around Go's HTTP server, to provide basic
command-line functionality, env-specific configuration, request logging, 
graceful shutdowns, and daemonization.

### Command Line

```Bash
$ myserver -h

Usage: myserver [options]

-a: ":9000"  Server address
-pid: "myserver.pid" Server PID File
-c: "myserver.cfg"  Config file
-e: "dev" Environment to run server in
-d: false Run server as daemon
-stop: false  Stop running server and exit
-restart: false Stop running server and boot daemon
-kill: false Force kill running server and exit
```


### Config File

```ini
[DEFAULT]
addr=":9000"
pidFile=path/to/file.pid
readTimeout=5s
writeTimeout=500ms
certFile=path/to/myserver.cert
keyFile=path/to/myserver.key

timeFormat=(02/01/2006 15:04:05)
logFormat=$RemoteAddr - $RemoteUser $Time "$Request" $Status $BodyBytes
logFile=path/to/myserver.log

customThing=foobar

[dev]
customThing=baz

[prod]
readTimeout=2s

```


### Go Code


```Go
package main

import (
  "fmt"
  "gosrv"
  "net/http"
)


func main() {
  s, err := gosrv.NewFromFlag()
  if err != nil { panic(err) }

  customThing, err := s.Config.String("customThing")
  if err != nil { fmt.Println("Custom thing is "+customThing+" for environment "+s.Config.Env) }

  s.HandleFunc("/", handler)

  err = s.ListenAndServe()
  if err != nil { panic(err) }
}


func handler(w http.ResponseWriter, r *http.Request) {
  w.Write([]byte("Hello World!\n"))

  res := w.(*gosrv.Response)

  for i := 0; i < 100; i++ {
    // Some long running request
    time.Sleep(50 * time.Millisecond)

    // Check to see if the server is trying to shut down
    if res.StopSignal() {
      break
    }
  }
}
```
