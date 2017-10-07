package common

import (
  "io"
  "fmt"
  "net"
  "./util"
)

type Connection struct {
  conn net.Conn
  connected bool
  id string
  read uint64
  listener *Server
}

type Server struct {
  listener net.Listener
  connected bool
}

const packet_size = 1024

func Create_Server(port string, conn_handler func(*Connection)) {
  ln, err := net.Listen("tcp", port)
  defer ln.Close()

  server := &Server {
    listener: ln,
    connected: true,
  }

  if err == nil {
    for server.connected {
      conn, err := ln.Accept()

      id := util.Get_Random_Id()
      if err == nil {
        connection := &Connection{
          conn: conn,
          connected: true,
          id: id,
          listener: server,
        }

        go conn_handler(connection)
      }
    }
  }
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

func (this * Server) Close() {
  this.connected = false
  this.listener.Close()
}

func (this * Connection) Connected() bool {
  return this.connected
}

func (this * Connection) Close() {
  this.conn.Close()
}

func (this * Connection) Id() string {
  return this.id
}

func (this * Connection) Write(m * util.Message) {
  this.conn.Write(m.Serialize())
}

func (this * Connection) Listener() * Server {
  return this.listener
}

func (this * Connection) Receive() * util.Message {
  // TODO (jkk111): Currently can't handle long message,
  // Awaiting spec to see if needed.
  buf := make([]byte, packet_size)
  read, err := this.conn.Read(buf)

  if err != nil {
    this.connected = false
    handle_conn_err(err)
    return nil
  } else {
    return util.Parse_Message(buf[:read])
  }
}

func Connect(url string) * Connection {
  conn, err := net.Dial("tcp", url)

  if err == nil {
    return &Connection{
      conn: conn,
      connected: true,
      read: 0,
      id: util.Get_Random_Id(),
    }
  } else {
    return nil
  }
}