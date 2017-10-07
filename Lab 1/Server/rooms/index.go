package rooms

import (
  "../../Common/util"
)

var rooms = make(map[string]*ChatRoom)

type ChatRoom struct {
  id string
  clients map[string]*ChatClient
}

type ChatClient struct {
  id string
  room * ChatRoom
}

func (this * ChatClient) join(room string) {

}

func Create_Room() * ChatRoom {
  id := util.Get_Random_Id()

  return &ChatRoom {
    id: id,
    clients: make(map[string]*ChatClient),
  }
}

func Rooms() map[string]*ChatRoom {
  return rooms
}