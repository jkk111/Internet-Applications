package main

/*
 * This is intended to be simple server to handle all of the requests
 * Ideally we would agree an id with a seperate messenger server,
 * and return the id to the conencted client
 * However in this instance:
 * When a client connects an id is immediately generated from the system random
 * source.
 * This id is then returned to the client, and we expect the client to reconnect
 * with the provided id.
 */


// Golang imports all native
import "os"
import "io"
import "net"
import "fmt"
import "time"
import "strconv"
import "crypto/rand"
import "encoding/hex"

// Define known Constants
const message_size = 1024;
const port = 8888
const random_id_bytes = 8


// Define required variables (rooms, logs, message_types)
var rooms map[string]*ChatRoom
var logs []*ActivityLog = make([]*ActivityLog, 0) // Define empty string array
var ctrl_sequence = map[string]string {
  "client_hello": "HELO",
  "client_message": "MESS",
  "client_disco": "DISCO",
  "client_join": "JOIN",
  "client_kill": "KILL",
} // Will update when spec is released

type ActivityLog struct {
  log_type string
  content string
  client string
  timestamp int64
} // Log Entry

type ChatRoom struct {
  id string
  clients map[string]*ChatClient
} // Represents a chat room and manages a hashmap of connected clients

type ChatClient struct {
  id string
  room *ChatRoom // Unsure if multiple rooms supported assuming 1 for now
  conn net.Conn
} // Represents a connected client and contains the socket for the client

type Message struct {
  message_type string
  message_body string
  m_cache []byte // Kind of expensive to convert to byte array, so cache result
} // Represents a message Object

type ChatMessage struct {
  sender string
  message Message
} // Represents a chat message.

func (this * Message) Serialize() []byte {
  if this.m_cache == nil {
    this.m_cache = []byte(this.message_type + " " + this.message_body)
  }
  return this.m_cache
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
} // Helper function to return a hex id

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
} // Creates and appends a log entry to the log list

func create_room() *ChatRoom {
  id := get_random_id()

  return &ChatRoom{
    id: id,
    clients: make(map[string]*ChatClient),
  }
} // Helper function to create a room

func (this * ChatClient) send(message Message) {
  this.conn.Write(message.Serialize())
}

func (this * ChatRoom) join(client * ChatClient) {
  this.clients[client.id] = client
}

func (this * ChatClient) join(room string) {
  if rooms[room] != nil {
    this.room = rooms[room]
    rooms[room].join(this)
  }
} // Adds a client to a chat room

func (this * ChatRoom) leave(client * ChatClient) {
  delete(this.clients, client.id)
} // Removes a client from a rooms client list

func (this * ChatClient) leave() {
  this.room.leave(this)
  this.room = nil
} // Removes the client from its active room

func (this * ChatRoom) send_message(message Message) {
  for _, client := range this.clients {
    client.conn.Write(message.Serialize())
  }
} // Sens a message to all clients in a room

// Returns a parsed message.
func parse_message(message []byte) Message {
  m_type := get_message_type(message)
  slice := len(m_type) + 1
  if slice > len(message) {
    slice = len(message)
  }
  return Message{
    message_type: m_type,
    message_body: string(message[slice:]),
    m_cache: message,
  }
}

// Gets the type of a message
// (or if invalid returns a string of the whole message)
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


// Handles an incoming message.
func handle_message(client * ChatClient, message []byte) {
  log("client_message", string(message), client)
  fmt.Printf("Read %d bytes, msg: %s\n", len(message), string(message))
  parsed := parse_message(message)
  switch parsed.message_type {
    case ctrl_sequence["client_hello"]:

      // Identify the client
      resp := Message{
        message_type: "IDENT",
        message_body: client.id,
      }
      client.send(resp)

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

// Deals with a connection error
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

// Watches a connection for incoming messages
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


// Entry Point
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