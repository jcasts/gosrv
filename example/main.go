package main

import (
  "net/http"
  "fmt"
  "os"
  "time"
  "../../gosrv"
)

func main() {
  s, err := gosrv.NewServerFromFlag()

  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  s.HandleFunc("/", func(wr http.ResponseWriter, req *http.Request){
    wr.Write([]byte("Hello World!\n"))
  })

  s.OnStop(func()error{
    fmt.Println("Sleeping for 1s")
    time.Sleep(1 * time.Second)
    return nil
  })

  panic( s.ListenAndServe() )
}
