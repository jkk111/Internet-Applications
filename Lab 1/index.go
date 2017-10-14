package main

import (
  "fmt"
  "net"
  "os"
  "io"
  "sync"
  "strconv"
)

const packet_size = 8192

var client_ids = Create_Counter()
var room_ids = Create_Counter()

var connected_clients = make(map[int]*Connection)

var rooms = make(map[int]*Room)
var rooms_mapped = make(map[string]*Room)

func Get_Room_By_Name(name string) * Room {
  if rooms_mapped[name] == nil {
    rooms_mapped[name] = &Room{
      room_ids.Next(),
      name,
      make(map[int]*Connection),
    }
  }

  return rooms_mapped[name]
}

func Get_Room_By_Id(id int) * Room {
  return rooms[id]
}

type counter struct {
  *sync.Mutex
  next_id int
}

func (this * counter) Next() int {
  this.Lock()
  id := this.next_id
  this.next_id++
  this.Unlock()
  return id
}

func Create_Counter() * counter {
  return &counter{
    &sync.Mutex{},
    1,
  }
}

type MessageComponent struct {
  Key, Value string
}

type Room struct {
  id int
  name string
  clients map[int]*Connection
}

func (this * Room) Join(conn * Connection) {
  this.clients[conn.id] = conn
}

func leave_message(room * Room, client string) * Message {
  message := fmt.Sprintf("%s has left this chatroom.\n\n", client)

  return Construct_Message([]*MessageComponent {
    Message_Component("CHAT:", fmt.Sprintf("%d", room.id)),
    Message_Component("CLIENT_NAME:", client),
    Message_Component("MESSAGE:", message),
  })
}

func (this * Room) Leave(conn * Connection, name string) {
  delete(this.clients, conn.id)
  this.Send(leave_message(this, name))
}

func (this * Room) Send(message * Message) {
  for _, client := range this.clients {
    client.Send(message)
  }
}

func (this * MessageComponent) len() int {
  term := 1
  if this.Key == "MESSAGE:" {
    term = 2
  }

  return len(this.Key) + 1 + len(this.Value) + term
}

func Construct_Message(components []*MessageComponent) * Message {
  return &Message{
    components: components,
  }
}

type Message struct {
  components []*MessageComponent
  mapped_components map[string]string
  m_cache []byte
}

func (this * Message) Type() string {
  return this.components[0].Key
}

func (this * Message) Value() string {
  return this.components[0].Value
}

func (this * Message) Serialize() []byte {
  if this.m_cache == nil {
    str := ""
    for _, component := range this.components {
      str += component.Key + " " + component.Value + "\n"
      if component.Key == "MESSAGE:" {
        str += "\n"
      }
    }
    this.m_cache = []byte(str)
  }

  return this.m_cache
}

func found(message []byte, index int, pattern []byte) bool {
  for i, c := range pattern {
    if message[i + index] != c {
      return false
    }
  }

  return true
}

// Gets the type of a message
// (or if invalid / type only, returns a string of the whole message)
func get_message_type(message []byte) string {
  // Strategy here is to read until the first space, buffer is limited to 1024
  // bytes so not too bad!

  for i := 0; i < len(message); i++ {
    if message[i] == ' ' || message[i] == '\n' {
      return string(message[:i])
    }
  }

  return string(message)
}

func parse_component_data(message []byte, terminator string) string {
  term_bytes := []byte(terminator)
  i := 0

  for ; i < len(message); i++ {
    if found(message, i, term_bytes) {
      break
    }
  }

  fmt.Println("MSG:", string(message[:i]), term_bytes)

  return string(message[:i])
}

func parse_component(message []byte) * MessageComponent {
  m_type := get_message_type(message)

  len_type := len(m_type) + 1

  term := "\n"
  if m_type == "MESSAGE:" {
    term = "\n\n"
  }

  if len_type > len(message) {
    return nil
  }

  component := parse_component_data(message[len_type:], term)

  return &MessageComponent{
    m_type,
    component,
  }
}

func Parse_Message(message []byte) * Message {
  components := make([]*MessageComponent, 0)

  first := parse_component(message)
  if first == nil {
    return nil
  }
  components = append(components, first)

  if first.len() < len(message) {
    message = message[first.len():]
  } else {
    message = make([]byte, 0)
  }

  for len(message) > 0 {
    component := parse_component(message)
    if component != nil {
      components = append(components, component)
      message = message[component.len():]
    } else {
      break
    }
  }

  mapped_components := make(map[string]string)

  for _, component := range components {
    mapped_components[component.Key] = component.Value
  }

  return &Message{
    components: components,
    mapped_components: mapped_components,
    m_cache: message,
  }
}

type Connection struct {
  id int
  name string // Last known name for client
  conn net.Conn
  connected bool
  rooms map[int]*Room
}

// Deals with a connection error
func handle_conn_err(err error) {
  if err == io.EOF {
    fmt.Println("Connection went away")
  } else if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
    fmt.Println("Connection Timeout")
  } else if operr, ok := err.(*net.OpError); ok {
    if operr.Op == "dial" {
      fmt.Println("Couldn't reach host")
    } else if operr.Op == "read" {
      fmt.Println("Can't read closed connection")
    } else {
      fmt.Printf("Failed to perform op: '%s'\n", operr.Op)
    }
  }
}

func Message_Component(key, value string) *MessageComponent {
  return &MessageComponent{ key, value }
}

func (this * Connection) Send(message * Message) {
  this.conn.Write(message.Serialize())
}

func (this * Connection) Close() {
  for _, room := range this.rooms {
    room.Leave(this, this.name)
  }
  this.conn.Close()
}

func (this * Connection) Receive() * Message {
  buf := make([]byte, packet_size)
  read, err := this.conn.Read(buf)

  if err != nil {
    this.connected = false
    delete(connected_clients, this.id)
    handle_conn_err(err)
    return nil
  } else {
    return Parse_Message(buf[:read])
  }
}

func client_hello(message * Message) * Message {
  return Construct_Message([]*MessageComponent{
    Message_Component("HELO", message.mapped_components["HELO"]),
    Message_Component("IP:", "0"),
    Message_Component("PORT:", "0"),
    Message_Component("StudentID:", "14319833"),
  })
}

func join_chatroom(conn * Connection, message * Message) * Message {
  room_name := message.mapped_components["JOIN_CHATROOM:"]
  room := Get_Room_By_Name(room_name)
  room.Join(conn)

  conn.rooms[room.id] = room

  room_id := fmt.Sprintf("%d", room.id)
  conn_id := fmt.Sprintf("%d", conn.id)

  return Construct_Message([]*MessageComponent {
    Message_Component("JOINED_CHATROOM", room_name),
    Message_Component("SERVER_IP", "0"),
    Message_Component("PORT", "0"),
    Message_Component("ROOM_REF", room_id),
    Message_Component("JOIN_ID", conn_id),
  })
}

func chat_message(conn * Connection, message * Message) {
  mapped := message.mapped_components
  room_id, _ := strconv.Atoi(mapped["CHAT:"])

  msg := Construct_Message([]*MessageComponent {
    Message_Component("CHAT:", mapped["CHAT:"]),
    Message_Component("CLIENT_NAME:", mapped["CLIENT_NAME:"]),
    Message_Component("MESSAGE:", mapped["MESSAGE:"]),
  })

  room := conn.rooms[room_id]
  room.Send(msg)
}

func leave_chatroom(conn * Connection, message * Message) * Message {
  mapped := message.mapped_components
  room_id, _ := strconv.Atoi(mapped["LEAVE_CHATROOM:"])
  name := mapped["CLIENT_NAME:"]

  conn.rooms[room_id].Leave(conn, name)
  delete(conn.rooms, room_id)

  return Construct_Message([]*MessageComponent {
    Message_Component("LEFT_CHATROOM:", mapped["LEAVE_CHATROOM:"]),
    Message_Component("JOIN_ID:", mapped["JOIN_ID:"]),
  })
}

func handle_message(conn * Connection, message * Message) {
  if message.mapped_components["CLIENT_NAME:"] != "" {
    conn.name = message.mapped_components["CLIENT_NAME:"]
  }

  switch message.Type() {
    case "HELO":
      conn.Send(client_hello(message))
      break
    case "JOIN_CHATROOM:":
      conn.Send(join_chatroom(conn, message))
      break
    case "CHAT:":
      chat_message(conn, message)
      break
    case "KILL_SERVICE":
      for _, client := range connected_clients {
        client.Close()
      }
      os.Exit(0)
      break
    case "DISCONNECT:":
      conn.Close()
      break
    case "LEAVE_CHATROOM:":
      conn.Send(leave_chatroom(conn, message))
      break
  }
}

func handle_conn(conn * Connection) {
  for conn.connected {
    message := conn.Receive()
    if conn.connected && message != nil {
      handle_message(conn, message)
    }
  }
}

func on_connect(c net.Conn) {
  conn := &Connection{
    client_ids.Next(),
    "",
    c,
    true,
    make(map[int]*Room),
  }

  connected_clients[conn.id] = conn

  handle_conn(conn)
}

func main() {
  port := ":" + os.Args[1]

  ln, err := net.Listen("tcp", port)

  if err != nil {
    fmt.Println("Failed to Listen on port", port)
    panic(err)
  }

  for {
    conn, err := ln.Accept()

    if err != nil {
      fmt.Println("Failed to get connection")
      panic(err)
    } else {
      go on_connect(conn)
    }
  }
}