package main

import "fmt"
import "net"
import "time"
import "strconv"

const port = 8888

func main() {
  tcp_location := "127.0.0.1:" + strconv.Itoa(port)
  conn, err := net.Dial("tcp",tcp_location)

  if err != nil {
    fmt.Printf("Failed to connect to the server at %s\n", tcp_location)
    panic(err)
  }

  defer conn.Close()

  conn.Write([]byte("HELO IAMHERE"))

  time.Sleep(time.Second)

  conn.Write([]byte("KILL"))

  resp := make([]byte, 1024)

  n, err := conn.Read(resp)

  if err != nil {
    fmt.Println("Failed to read the response")
  } else {
    fmt.Printf("Response: %s\n", string(resp[:n]))
  }
}