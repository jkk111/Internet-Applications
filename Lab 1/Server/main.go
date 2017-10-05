package main

import "net"
import "fmt"
import "time"
import "strconv"

const package_size = 1024;
const port = 8888

var rooms map[string]*ChatRoom

var logs []*ActivityLog = make([]*ActivityLog, 0) // Define empty string array

type ActivityLog struct {
  log_type string
  content string
  client string
  timestamp int64
}

type ChatRoom struct {
  id string
  clients []ChatClient
}

type ChatClient struct {
  id string
  room *ChatRoom // Unsure if multiple rooms supported assuming 1 for now
}

func log(log_type, content string, client * ChatClient) {
  client_id := ""
  if client != nil {
    client_id = client.id
  }
  entry := &ActivityLog{
    log_type,
    content,
    client_id,
    time.Now().Unix(),
  }
  logs = append(logs, entry)
}

func (this * ChatClient) join_room() {

}

func handle_message() {

}

func handle_connection(conn net.Conn) {

}

func main() {
  port_str := ":" + strconv.Itoa(port)
  ln, err := net.Listen("tcp", port_str)

  if err != nil {
    fmt.Println("Failed to listen on port %d", port)
    panic(err)
  }

  for {
    conn, err := ln.Accept()

    if err != nil {
      fmt.Println("Failed to accept client connection")
    } else {
      go handle_connection(conn)
    }
  }
}