package main

import (
  "../../Common"
  "../../Common/util"
  "../rooms"
  "fmt"
)

const port = ":8888"
const sid = "14319833"

func handle_message(conn * rooms.ChatClient, message * util.Message) {
  fmt.Printf("Recevied Message %s\n", message)
  switch message.Type() {
    case "HELO":

      response := []*util.MessageComponent {
        util.Message_Component("HELO", message.Body()["HELO"]),
        util.Message_Component("IP:", "0"),
        util.Message_Component("PORT:", "0"),
        util.Message_Component("StudentID", sid),
      }

      resp := util.Create_Message(response)
      conn.Write(resp)
      break

    case "KILL_SERVICE":
      ln := conn.Listener()
      ln.Close()
      break

    case "JOIN_CHATROOM:":
      conn.Join(message.Body())
      break

    case "CHAT:"
      conn.Chat(message.Body())
      break

    default:
      conn.Close()
  }
}

func conn_handler(conn * common.Connection) {
  client := &rooms.ChatClient{conn, nil}
  for client.Connected() {
    message := client.Receive()
    if message != nil {
      handle_message(client, message)
    }
  }
}

func main() {
  common.Create_Server(port, conn_handler)
}