package main

import (
  "fmt"
  "./Common"
)

func main() {
  client := common.Connect("127.0.0.1:8888")
  if client != nil {
    client.Write([]byte("HELO"))
    for client.Connected() {
      message := client.Receive()
      if message != nil {
        fmt.Println(message)
      }
    }
  }
}