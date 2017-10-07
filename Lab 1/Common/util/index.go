package util

import (
  "fmt"
  "encoding/hex"
  "crypto/rand"
)

const random_id_bytes = 8

type Message struct {
  message_type string
  message_body string
  m_cache []byte // Kind of expensive to convert to byte array, so cache result
} // Represents a message Object

func (this * Message) Type() string {
  return this.message_type
}

func (this * Message) Body() string {
  return this.message_body
}

func (this * Message) String() string {
  str := this.message_type
  if this.message_body != "" {
    str += " " + this.message_body
  }
  return str
}

func (this * Message) Equals(other * Message) bool {
  return (this.Type() == "IDENT" && other.Type() == "IDENT" ||
           (this.Type() == other.Type() &&
             this.Body() == other.Body()))
}

func (this * Message) Serialize() []byte {
  if this.m_cache == nil {
    str := this.Type()
    if this.message_body != "" {
      str = str + " " + this.message_body
    }
    this.m_cache = []byte(str)
  }
  return this.m_cache
}

// Returns a parsed message.
func Parse_Message(message []byte) * Message {
  m_type := get_message_type(message)
  slice := len(m_type) + 1
  if slice > len(message) {
    slice = len(message)
  }
  return &Message{
    message_type: m_type,
    message_body: string(message[slice:]),
    m_cache: message,
  }
}

func Create_Message(message_type, body string) * Message {
  return &Message{
    message_type: message_type,
    message_body: body,
  }
}

func Create_Message_From_String(message string) * Message {
  m_type := get_message_type([]byte(message))
  slice := len(m_type) + 1
  if slice > len(m_type) {
    slice = len(m_type)
  }

  return &Message {
    message_type: message[:slice],
    message_body: message[slice:],
  }
}

func Get_Random_Id() string {
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

// Gets the type of a message
// (or if invalid / type only, returns a string of the whole message)
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