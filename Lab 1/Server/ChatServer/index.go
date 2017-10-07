package main

import (
  "../../Common"
  "../../Common/util"
  "../rooms"
  "fmt"
)

var port = ":8888"

func handle_message(conn * common.Connection, message * util.Message) {
  fmt.Printf("Recevied Message %s\n", message)
  switch message.Type() {
    case "HELO":
      resp := util.Create_Message("IDENT", conn.Id())
      conn.Write(resp)
      break

    case "TESTING":
      resp := util.Create_Message("TESTING", "")
      conn.Write(resp)
      break

    case "KILL":
      ln := conn.Listener()
      ln.Close()
      break

    default:
      m := util.Create_Message("REJECT", "")
      conn.Write(m)
      conn.Close()
  }
}

func conn_handler(conn * common.Connection) {
  for conn.Connected() {
    message := conn.Receive()
    if message != nil {
      // Do something here
      handle_message(conn, message)
    }
  }
}

func main() {
  _ = rooms.Rooms // Silences error regarding unused package
  common.Create_Server(port, conn_handler)
}