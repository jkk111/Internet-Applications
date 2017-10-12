package rooms

import (
  "sync"
  "strconv"
  "../../Common"
)

type counter struct {
  sync.Mutex
  next_id int
}

func create_counter() *counter {
  return &counter{
    sync.Mutex{},
    1,
  }
}

func (this * counter) next() int {
  this.Lock()
  id := this.next_id
  this.next_id++
  this.Unlock()
  return id
}

var room_ids = create_counter()

var rooms = make([]*ChatRoom, 0)

var room_maps = make(map[string]int)

type ChatRoom struct {
  id int
  name string
  clients map[int]*ChatClient
}

func (this * ChatRoom) Leave(client * ChatClient) {
  delete(this.clients, client.id)
}

type ChatClient struct {
  *common.Connection
  Rooms map[int]*ChatRoom
}

func get_room(id int) *ChatRoom {
  return rooms[id - 1]
}

func room_id(name string) int {
  if room_maps[name] != 0 {
    return room_maps[name]
  } else {
    id := room_ids.next()
    room := &ChatRoom{
      id: id,
      name: name,
      clients: make(map[string]*ChatClient, 0),
    }
    room_maps[name] = id
    rooms[id - 1] = room
    return id
  }
}

func (this * ChatClient) Join(request map[string]string) {
  room_name := request["JOIN_CHATROOM:"]
  room := room_id(room_name)
  name := request["CLIENT_NAME:"]
}

func (this * ChatClient) Leave(request map[string]string) {
  room_id, _ := strconv.Atoi(request["LEAVE_CHATROOM:"])
  room := get_room(room_id)
  delete(this.Rooms, room_id)
  room.Leave(this)
}

func (this * ChatClient) Chat(request map[string]string) {
  room_id, _ := strconv.Atoi(request["CHAT:"])
  name := request["CLIENT_NAME:"]
  message := request["MESSAGE:"]

  room := get_room(room_id)
}