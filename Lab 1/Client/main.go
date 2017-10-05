package main

import "fmt"
import "net"
import "strconv"

const port = 8888

func main() {
  tcp_location := "127.0.0.1:" + strconv.Itoa(port)
  conn, err := net.Dial("tcp",tcp_location)

  if err != nil {
    fmt.Println("Failed to connect to the server at %s", tcp_location)
    panic(err)
  }

  defer conn.Close()

  conn.Write([]byte("Hello world"))

  resp := make([]byte, 1024)

  n, err := conn.Read(resp)

  if err != nil {
    fmt.Println("Failed to read the response")
  } else {
    fmt.Printf("Response: %s\n", string(resp[:n]))
  }
}