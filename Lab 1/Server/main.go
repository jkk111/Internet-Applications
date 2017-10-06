package main

import "os"
import "io"
import "net"
import "fmt"
import "time"
import "strconv"
import "crypto/rand"
import "encoding/hex"

const message_size = 1024;
const port = 8888
const random_id_bytes = 8

var rooms map[string]*ChatRoom

var logs []*ActivityLog = make([]*ActivityLog, 0) // Define empty string array

var ctrl_sequence = map[string]string {
  "client_hello": "HELO",
  "client_message": "MESS",
  "client_disco": "DISCO",
  "client_join": "JOIN",
  "client_kill": "KILL",
}

type ActivityLog struct {
  log_type string
  content string
  client string
  timestamp int64
}

type ChatRoom struct {
  id string
  clients map[string]*ChatClient
}

type ChatClient struct {
  id string
  room *ChatRoom // Unsure if multiple rooms supported assuming 1 for now
  conn net.Conn
}

func get_random_id() string {
  buf := make([]byte, random_id_bytes)
  _, err := rand.Read(buf)
  if err != nil {
    fmt.Println("Failed to read from random source")
    panic(err)
  }

  encodedLen := hex.EncodedLen(len(buf))
  hexDest := make([]byte, encodedLen)
  hex.Encode(hexDest, buf)
  return string(hexDest)
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

func create_room() *ChatRoom {
  id := get_random_id()

  return &ChatRoom{
    id: id,
    clients: make(map[string]*ChatClient),
  }
}

func (this * ChatClient) Send(m_type, message string) {
  temp := m_type + " " + message

  this.conn.Write([]byte(temp))
}

func (this * ChatRoom) join(client * ChatClient) {
  this.clients[client.id] = client
}

func (this * ChatClient) join(room string) {
  this.room = rooms[room]
  rooms[room].join(this)
}

func (this * ChatRoom) leave(client * ChatClient) {
  delete(this.clients, client.id)
}

func (this * ChatClient) leave() {
  this.room.leave(this)
  this.room = nil
}

func (this * ChatRoom) send_message(message, sender string) {
  for _, client := range this.clients {
    client.conn.Write([]byte(message))
  }
}

func get_message_type(message []byte) string {
  // Strategy here is to read until the first space, buffer is limited to 1024
  // bytes so not too bad!

  for i := 0; i < len(message); i++ {
    if message[i] == ' ' {
      return string(message[:i])
    }
  }

  return string(message)
}

func handle_message(client * ChatClient, message []byte) {
  log("client_message", string(message), client)
  fmt.Printf("Read %d bytes, msg: %s\n", len(message), string(message))
  message_type := get_message_type(message)
  switch message_type {
    case ctrl_sequence["client_hello"]:

      // Identify the client

      client.Send("IDENT", client.id)

      break
    case ctrl_sequence["client_disco"]:
      break
    case ctrl_sequence["client_message"]:
      break
    case ctrl_sequence["client_kill"]:
      os.Exit(0)
    default:
      client.conn.Write([]byte("Error Unknown Request!"))
      client.conn.Close()
  }
}

func handle_conn_err(err error) {
  fmt.Printf("%T %+v\n", err, err)
  if err == io.EOF {
    fmt.Println("Connection went away")
  } else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
    fmt.Println("Connection Timeout")
  } else if operr, ok := err.(*net.OpError); ok {
    if operr.Op == "dial" {
      fmt.Println("Couldn't reach host")
    } else if operr.Op == "read" {
      fmt.Println("Can't write to closed connection")
    }
  }
}

func handle_connection(conn net.Conn) {
  rand_id := get_random_id()

  client := &ChatClient{
    id: rand_id,
    conn: conn,
  }

  connected := true

  for connected {
    buf := make([]byte, message_size)
    n, err := conn.Read(buf)
    buf = buf[:n]
    if err != nil {
      log("error", err.Error(), client)
      handle_conn_err(err)
      connected = false
    } else {
      handle_message(client, buf)
    }
  }
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