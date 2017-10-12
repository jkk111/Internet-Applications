package util

import (
  "fmt"
  "encoding/hex"
  "crypto/rand"
)

const random_id_bytes = 8

type Message struct {
  components map[string]string
  message_type string
  message_body string
  m_cache []byte // Kind of expensive to convert to byte array, so cache result
} // Represents a message Object

type MessageComponent struct {
  component_type string
  data string
  len int
}

func Message_Component(c_type, data string) *MessageComponent {
  return &MessageComponent{
    component_type: c_type,
    data: data,
  }
}

func (this * Message) Type() string {
  return this.message_type
}

func (this * Message) Body() map[string]string {
  return this.components
}

func (this * Message) String() string {
  str := ""
  for key, comp := range this.components {
    str += key + " " + comp + "\n"
    if key == "MESSAGE:" {
      str += "\n"
    }
  }
  return str
}

func (this * Message) Equals(other * Message) bool {
  if this.Type() != other.Type() {
    return false
  }

  for key, entry := range this.components {
    if entry != other.components[key] {
      return false
    }
  }

  for key, entry := range other.components {
    if entry != this.components[key] {
      return false
    }
  }

  return true
}

func (this * Message) Serialize() []byte {
  if this.m_cache == nil {
    if this.message_body != "" {
      this.m_cache = []byte(this.message_body)
    } else {
      str := ""

      fmt.Println(this.components)

      for key, c := range this.components {
        str += key + " " + c + "\n"
        if key == "MESSAGE:" {
          str += "\n"
        }
      }

      fmt.Println("Serialized", len(str))

      this.m_cache = []byte(str)
    }
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

func parse_component_data(message []byte, terminator string) string {
  term_bytes := []byte(terminator)
  i := 0
  for ; i < len(message); i++ {
    if found(message, i, term_bytes) {
      break
    }
  }

  return string(message[:i])
}

func parse_component(message []byte) *MessageComponent {
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
    len_type + len(component) + len(term),
  }
}

// Returns a parsed message.
func Parse_Message(message []byte) * Message {
  components := make([]*MessageComponent, 0)

  first := parse_component(message)
  components = append(components, first)
  fmt.Println(first, message, first.len)
  if first.len < len(message) {
    message = message[first.len:]
  } else {
    message = make([]byte, 0)
  }

  for len(message) > 0 {
    component := parse_component(message)
    if component != nil {
      components = append(components, component)
      message = message[component.len:]
    } else {
      break
    }
  }

  mapped_components := make(map[string]string)

  for _, component := range components {
    mapped_components[component.component_type] = component.data
  }

  return &Message{
    components: mapped_components,
    message_type: first.component_type,
    m_cache: message,
  }
}

func Create_Message(data []*MessageComponent) * Message {
  content := ""

  for _, item := range data {
    content += item.component_type + " " + item.data + "\n"
    if item.component_type == "MESSAGE:" {
      content += "\n"
    }
  }

  return &Message{
    message_body: content,
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