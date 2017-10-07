package main

import (
  "../../Common"
  "../../Common/util"
  "fmt"
)

var port = ":8888"

func handle_message(conn * common.Connection, message * util.Message) {
  fmt.Printf("Recevied Message %s\n", message)
  switch message.Type() {
    default:
      m := util.Create_Message("Reject", "")
      conn.Write(m)
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
  common.Create_Server(port, conn_handler)
}