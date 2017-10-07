package main

/*
 * Client to be used for automated testing
 */

import "os"
import "fmt"
import "bufio"
import "strings"
import "strconv"
import "../Common"
import "../Common/util"

type ClientMessage struct {
  message * util.Message
  outbound bool
}

const port = 8888

func read_conversations() [][]*ClientMessage {
  f, err := os.Open("./Messages.txt")
  if err != nil {
    fmt.Println("Failed to open messages file")
    panic(err)
  }

  defer f.Close()
  scanner := bufio.NewScanner(f)
  conversations := make([][]*ClientMessage, 0)
  conversation := make([]*ClientMessage, 0)

  for scanner.Scan() {
    line := scanner.Text()
    if strings.Trim(line, " \n\t") == "" {
      if len(conversation) > 0 {
        conversations = append(conversations, conversation)
        conversation = make([]*ClientMessage, 0)
      }
      continue
    }

    var message * util.Message
    message = util.Create_Message_From_String(line[1:])
    outbound := line[0] == '>'

    m := &ClientMessage {
      message: message,
      outbound: outbound,
    }

    conversation = append(conversation, m)
  }

  if len(conversation) > 0 {
    conversations = append(conversations, conversation)
  }

  return conversations
}

func main() {
  tcp_location := "127.0.0.1:" + strconv.Itoa(port)
  conversations := read_conversations()
  for test_index, conversation := range conversations {
    res_index := 0
    conn := common.Connect(tcp_location)
    if conn != nil {
      defer conn.Close()

      for _, message := range conversation {
        if message.outbound {
          conn.Write(message.message)
        } else {
          m := conn.Receive()

          if !m.Equals(message.message) {
            fmt.Printf("%s != %s\n", message.message, m)
            panic("Assertion Error!")
          } else {
            fmt.Printf("(%d:%d) Message matches expected response\n", test_index,
                                                                      res_index)
            res_index++
          }
        }
      }
    } else {
      fmt.Println("Failed to connect to the remote server")
    }
  }
}