package main

import (
  "net/http"
  "fmt"
  "../../gosrv"
)

func main() {
  ch := make(chan bool, 5)

  for i := 0; i < 5; i++ {
    s := gosrv.New()
    s.Addr = fmt.Sprintf(":900%d", i)
    s.HandleFunc("/", handler)

    go func(){
      s.ListenAndServe()
      ch <- true
    }()
  }

  for i := 0; i < 5; i++ {
    <-ch
  }
}


func handler(wr http.ResponseWriter, req *http.Request) {
  wr.Write([]byte("Hello World!\n"))
}
