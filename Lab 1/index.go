package main

import (
  "fmt"
  "net"
  "os"
  "io"
  "sync"
  "time"
  "strconv"
  "strings"
)

const packet_size = 8192
const StudentID = "14319833"

var client_ids = Create_Counter()
var room_ids = Create_Counter()

var E_NO_ROOM = Construct_Message([]*MessageComponent {
  Message_Component("ERROR_CODE:", "1"),
  Message_Component("ERROR_DESCRIPTION:", "Error: Cannot connect to non-existant room"),
})

var E_NO_FUNC = Construct_Message([]*MessageComponent {
  Message_Component("ERROR_CODE:", "2"),
  Message_Component("ERROR_DESCRIPTION:", "Error: Invalid Method"),
})

var connected_clients = make(map[int]*Connection)
var connected_clients_mutex = sync.Mutex{}

var rooms = make(map[int]*Room)
var rooms_mapped = make(map[string]*Room)

// Chat Server Functions

func client_hello(text, addr, port string) * Message {
  return Construct_Message([]*MessageComponent{
    Message_Component("HELO", text),
    Message_Component("IP:", addr),
    Message_Component("PORT:", port),
    Message_Component("StudentID:", StudentID),
  })
}

func joined_room(room_id, client_name string) * Message {
  msg := fmt.Sprintf("%s has joined this chatroom.", client_name)
  return Construct_Message([]*MessageComponent {
    Message_Component("CHAT:", room_id),
    Message_Component("CLIENT_NAME:", client_name),
    Message_Component("MESSAGE:", msg),
  })
}

func join_room(room_name, client_name, ip, port, join_id string) * Message {
  room := Get_Room_By_Name(room_name)
  return Construct_Message([]*MessageComponent {
    Message_Component("JOINED_CHATROOM:", room_name),
    Message_Component("SERVER_IP:", ip),
    Message_Component("PORT:", port),
    Message_Component("ROOM_REF:", room.id),
    Message_Component("JOIN_ID:", join_id),
  })
}



func leave_message(room * Room, client string) * Message {
  message := fmt.Sprintf("%s has left this chatroom.", client)

  return Construct_Message([]*MessageComponent {
    Message_Component("CHAT:", room.id),
    Message_Component("CLIENT_NAME:", client),
    Message_Component("MESSAGE:", message),
  })
}

func leave_room(ref, join_id string) * Message {
  return Construct_Message([]*MessageComponent{
    Message_Component("LEFT_CHATROOM:", ref),
    Message_Component("JOIN_ID:", join_id),
  })
}

func chat_message(ref, client_name, message string) * Message {
  room_id, _ := strconv.Atoi(ref)
  room := rooms[room_id]

  if room == nil {
    return E_NO_ROOM
  } else {
    msg := Construct_Message([]*MessageComponent {
      Message_Component("CHAT:", ref),
      Message_Component("CLIENT_NAME:", client_name),
      Message_Component("MESSAGE:", message),
    })
    room.Send(msg)
    return nil
  }
}

// End chat server functions

func Get_Room_By_Name(name string) * Room {
  if rooms_mapped[name] == nil {
    id := room_ids.Next()
    room := &Room{
      id,
      name,
      make(map[int]*Connection),
    }
    room_id, _ := strconv.Atoi(id)
    rooms[room_id] = room
    rooms_mapped[name] = room
  }

  return rooms_mapped[name]
}

func Get_Room_By_Id(id int) * Room {
  fmt.Println(id, rooms)
  return rooms[id]
}

type counter struct {
  *sync.Mutex
  next_id int
}

func (this * counter) Next() string {
  this.Lock()
  id := this.next_id
  this.next_id++
  this.Unlock()
  return strconv.Itoa(id)
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
  id string
  name string
  clients map[int]*Connection
}

func (this * Room) Join(conn_id int) {
  this.clients[conn_id] = connected_clients[conn_id]
}

func (this * Room) Leave(join_id, name string) {
  conn_id, _ := strconv.Atoi(join_id)
  this.Send(leave_message(this, name))
  delete(this.clients, conn_id)
}


func (this * Room) Send(message * Message) {
  fmt.Println(len(this.clients))
  for _, client := range this.clients {
    client.Send(message)
  }
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
  m_type string
}

func (this * Message) Type() string {
  // return this.components[0].Key
  return this.m_type
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
    if message[i] == ' ' || message[i] == '\n' || message[i] == ':' {
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
  if i == len(message) {
    return string(message)
  }
  return string(message[:i])
}

func seek(message []byte, start int, c byte) int {
  i := start
  for i < len(message) && message[i] == c {
    i++
  }
  return i
}

func Parse_Messages(conn * Connection, message []byte) []*Message {
  messages := make([]*Message, 0)

  read := 0
  l := len(message)
  last := 0
  for read < l {
    m_type := get_message_type(message[read:])
    read += len(m_type) + 1
    read = seek(message, read, ' ')
    m_data := make(map[string]string)
    switch m_type {
      case "HELO":
        ident := parse_component_data(message[read:], "\n")
        m_data["HELO"] = ident
        read += len(ident) + 1
        break
      case "JOIN_CHATROOM":
        room := parse_component_data(message[read:], "\n")
        read += len(room) + 1 + len("CLIENT_IP:")
        read = seek(message, read, ' ')
        ip := parse_component_data(message[read:], "\n")
        read += len(ip) + 1 + len("PORT:")
        read = seek(message, read, ' ')
        port := parse_component_data(message[read:], "\n")
        read += len(port) + 1 + len("CLIENT_NAME:")
        read = seek(message, read, ' ')
        c_name := parse_component_data(message[read:], "\n")
        read += len(c_name) + 1
        m_data["JOIN_CHATROOM:"] = room
        m_data["CLIENT_NAME:"] = c_name
        m_data["PORT:"] = port
        m_data["CLIENT_IP:"] = ip
        break
      case "LEAVE_CHATROOM":
        room := parse_component_data(message[read:], "\n")
        read += len(room) + 1 + len("JOIN_ID:")
        read = seek(message, read, ' ')
        id := parse_component_data(message[read:], "\n")
        read += len(id) + 1 + len("CLIENT_NAME:")
        read = seek(message, read, ' ')
        c_name := parse_component_data(message[read:], "\n")
        read += len(c_name) + 1
        m_data["LEAVE_CHATROOM:"] = room
        m_data["JOIN_ID:"] = id
        m_data["CLIENT_NAME:"] = c_name
        break
      case "DISCONNECT":
        disco := parse_component_data(message[read:], "\n")
        read += len(disco) + 1 + len("PORT:")
        read = seek(message, read, ' ')
        port := parse_component_data(message[read:], "\n")
        read += len(port) + 1 + len("CLIENT_NAME:")
        read = seek(message, read, ' ')
        c_name := parse_component_data(message[read:], "\n")
        read += len(c_name) + 1
        m_data["DISCONNECT:"] = disco
        m_data["PORT:"] = port
        m_data["CLIENT_NAME:"] = c_name
        break
      case "CHAT":
        room := parse_component_data(message[read:], "\n")
        read += len(room) + 1 + len("JOIN_ID:")
        read = seek(message, read, ' ')
        id := parse_component_data(message[read:], "\n")
        read += len(id) + 1 + len("CLIENT_NAME:")
        read = seek(message, read, ' ')
        c_name := parse_component_data(message[read:], "\n")
        read += len(c_name) + 1 + len("MESSAGE:")
        read = seek(message, read, ' ')
        m := parse_component_data(message[read:], "\n\n")
        read += len(m) + 2
        m_data["CHAT:"] = room
        m_data["JOIN_ID:"] = id
        m_data["CLIENT_NAME:"] = c_name
        m_data["MESSAGE:"] = m
        break
      default:
        break
    }
    m := &Message{
      nil, m_data, message[last:read], m_type,
    }

    last = read
    messages = append(messages, m)
  }

  fmt.Printf("Parsed %d messages\n", len(messages))

  return messages
}

type Connection struct {
  id string
  name string // Last known name for client
  conn net.Conn
  connected bool
  log * os.File
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
  fmt.Println("SENDING", string(message.Serialize()))
  l, _ := this.conn.Write(message.Serialize())

  fmt.Printf("Wrote %d bytes to socket\n", l)
}

func (this * Connection) Addr() (string, string) {
  addr := this.conn.LocalAddr().String()
  parts := strings.Split(addr, ":")
  ip := parts[0]
  port := parts[1]
  return ip, port
}

func (this * Connection) Close() {
  fmt.Println("Closing connection with", this.id)

  for _, room := range this.rooms {
    room.Leave(this.id, this.name)
  }
  this.conn.Close()

  this.connected = false
  connected_clients_mutex.Lock()
  conn_id, _ := strconv.Atoi(this.id)
  delete(connected_clients, conn_id)
  connected_clients_mutex.Unlock()
}

func (this * Connection) Receive() []*Message {
  buf := make([]byte, packet_size)
  read, err := this.conn.Read(buf)

  if err != nil {
    this.Close()
    handle_conn_err(err)
    return nil
  } else {
    this.log.Write(buf[:read])
    return Parse_Messages(this, buf[:read])
  }
}

func handle_message(conn * Connection, message * Message) {
  fmt.Printf("Handling: %s\n", message.Type())
  if message.mapped_components["CLIENT_NAME:"] != "" {
    conn.name = message.mapped_components["CLIENT_NAME:"]
  }

  switch message.Type() {
    case "HELO":
      text := message.mapped_components["HELO"]
      ip, port := conn.Addr()
      conn.Send(client_hello(text, ip, port))
      break
    case "JOIN_CHATROOM":
      room_name := message.mapped_components["JOIN_CHATROOM:"]
      client_name := conn.name
      ip, port := conn.Addr()
      join_id := conn.id
      conn.Send(join_room(room_name, client_name, ip, port, join_id))

      room := Get_Room_By_Name(room_name)
      conn_id, _ := strconv.Atoi(conn.id)
      room.Join(conn_id)
      room_id, _ := strconv.Atoi(room.id)
      conn.rooms[room_id] = room
      m := joined_room(room.id, client_name)
      room.Send(m)
      break
    case "CHAT":
      ref := message.mapped_components["CHAT:"]
      client_name := conn.name
      msg := message.mapped_components["MESSAGE:"]
      chat_message(ref, client_name, msg)
      break

    case "KILL_SERVICE":
      for _, client := range connected_clients {
        client.Close()
      }

      // Wait for connections to close gracefully
      for len(connected_clients) > 0 {
        time.Sleep(time.Millisecond)
      }

      os.Exit(0)
      break

    case "DISCONNECT":
      conn.log.Close()
      id, _ := strconv.Atoi(conn.id)
      fmt.Println(conn.name, conn.id, connected_clients, connected_clients[id].rooms)
      conn.Close()
      break

    case "LEAVE_CHATROOM":
      ref := message.mapped_components["LEAVE_CHATROOM:"]
      room_id, _ := strconv.Atoi(ref)
      conn.Send(leave_room(ref, conn.id))
      room := rooms[room_id]
      room.Leave(conn.id, conn.name)
      delete(conn.rooms, room_id)
      break
  }
}

// Handles incoming messages for a connection
func handle_conn(conn * Connection) {
  for conn.connected {
    messages := conn.Receive()
    if conn.connected && messages != nil {
      for _, message := range messages {
        fmt.Println("Received message", string(message.Serialize()))
        handle_message(conn, message)
      }
    }
  }
}

// Upgrades our raw tcp connection to one with some more client info
func create_connection(c net.Conn) * Connection {
  file_name := fmt.Sprintf("w_stream %d", time.Now().UnixNano())
  f, _ := os.Create(file_name)
  return &Connection{
    client_ids.Next(),
    "",
    c,
    true,
    f,
    make(map[int]*Room),
  }
}

// Handles a client connecting
func on_connect(c net.Conn) {
  conn := create_connection(c)
  connected_clients_mutex.Lock()
  conn_id, _ := strconv.Atoi(conn.id)
  connected_clients[conn_id] = conn
  connected_clients_mutex.Unlock()
  handle_conn(conn)
}

// Main Entry Point
func main() {
  port := ":" + os.Getenv("PORT")

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
      fmt.Printf("\n\n\nClient Connected\n\n\n")
      go on_connect(conn)
    }
  }
}