package main

import "time"

const package_size = 1024;

var rooms map[string]*ChatRoom

var logs []*ActivityLog = make([]*ActivityLog, 0) // Define empty string array

type ActivityLog struct {
  log_type string
  content string
  client string
  timestamp int64
}

type ChatRoom struct {
  id string
  clients []ChatClient
}

type ChatClient struct {
  id string
  room *ChatRoom // Unsure if multiple rooms supported assuming 1 for now
}

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
}

func (this * ChatClient) join_room() {

}

func handle_message() {

}

func handle_connection() {

}

func main() {

}