package main

import (
  "net/http"
  "fmt"
  "os"
  "io/ioutil"
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
    ioutil.WriteFile("done.txt", []byte("main finished"), 0666)
    fmt.Println("Sleeping for 1s")
    time.Sleep(1 * time.Second)
    return nil
  })

  panic( s.ListenAndServe() )
}