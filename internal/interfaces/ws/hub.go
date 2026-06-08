package ws

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type BroadcastMessage struct {
	Message []byte
	Topic   string
}

type Client struct {
	Channel          chan BroadcastMessage
	HeartbeatChannel chan []byte
	Topics           []string // was: Topic string
	UserID           uuid.UUID
	DeviceID         uuid.UUID
	IsActive         bool
}

type Hub struct {
	clientsByUser          map[uuid.UUID][]*Client
	clientsByDevice        map[uuid.UUID]*Client
	deviceByChannel        map[chan BroadcastMessage]uuid.UUID
	usersByTopic           map[string]map[uuid.UUID]struct{}
	lastBroadCastedMessage map[string]map[uuid.UUID]string
	Broadcast              chan BroadcastMessage
	Register               chan Client
	Unregister             chan Client
	Ping                   chan []byte
	shutdown               chan struct{}
	Done                   chan struct{}
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:              make(chan BroadcastMessage),
		Ping:                   make(chan []byte),
		Register:               make(chan Client),
		Unregister:             make(chan Client),
		clientsByUser:          make(map[uuid.UUID][]*Client),
		clientsByDevice:        make(map[uuid.UUID]*Client),
		deviceByChannel:        make(map[chan BroadcastMessage]uuid.UUID),
		usersByTopic:           make(map[string]map[uuid.UUID]struct{}),
		lastBroadCastedMessage: make(map[string]map[uuid.UUID]string),
		shutdown:               make(chan struct{}),
		Done:                   make(chan struct{}),
	}
}

func (h *Hub) Run() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.Ping <- []byte(`{"type":"ping"}`)
			case <-h.shutdown:
				return
			}
		}
	}()

	for {
		select {
		case client := <-h.Register:
			h.clientsByDevice[client.DeviceID] = &client
			h.deviceByChannel[client.Channel] = client.DeviceID
			h.clientsByUser[client.UserID] = append(h.clientsByUser[client.UserID], &client)

			for _, topic := range client.Topics {
				if _, ok := h.usersByTopic[topic]; !ok {
					h.usersByTopic[topic] = make(map[uuid.UUID]struct{})
				}
				h.usersByTopic[topic][client.UserID] = struct{}{}
			}

			fmt.Printf("Hub: User %s (device %s) connected to topics %v\n",
				client.UserID, client.DeviceID, client.Topics)

		case client := <-h.Unregister:
			h.unregisterDevice(client.DeviceID)

		case message := <-h.Ping:
			for deviceID, client := range h.clientsByDevice {
				select {
				case client.HeartbeatChannel <- message:
				default:
					h.cleanupDeadConnection(deviceID)
				}
			}

		case message := <-h.Broadcast:
			log.Printf("Broadcast topic: %s to all connected users", message.Topic)

			if _, ok := h.lastBroadCastedMessage[message.Topic]; !ok {
				h.lastBroadCastedMessage[message.Topic] = make(map[uuid.UUID]string)
			}

			for userID := range h.usersByTopic[message.Topic] {
				if h.lastBroadCastedMessage[message.Topic][userID] == string(message.Message) {
					continue
				}
				for _, client := range h.clientsByUser[userID] {
					select {
					case client.Channel <- message:
						h.lastBroadCastedMessage[message.Topic][userID] = string(message.Message)
					default:
						h.cleanupDeadConnection(client.DeviceID)
					}
				}
			}

		case <-h.shutdown:
			h.gracefulShutdown()
			return
		}
	}
}

// unregisterDevice centralises the removal logic used by both Unregister and
// cleanupDeadConnection so the two paths stay consistent.
func (h *Hub) unregisterDevice(deviceID uuid.UUID) {
	client, ok := h.clientsByDevice[deviceID]
	if !ok {
		return
	}

	// Remove from user's device list.
	userClients := h.clientsByUser[client.UserID]
	for i, c := range userClients {
		if c.DeviceID == deviceID {
			h.clientsByUser[client.UserID] = append(userClients[:i], userClients[i+1:]...)
			break
		}
	}

	// If user has no devices left, remove all topic subscriptions.
	if len(h.clientsByUser[client.UserID]) == 0 {
		delete(h.clientsByUser, client.UserID)

		for _, topic := range client.Topics {
			if topicUsers, ok := h.usersByTopic[topic]; ok {
				delete(topicUsers, client.UserID)
				if len(topicUsers) == 0 {
					delete(h.usersByTopic, topic)
				}
			}
		}
	}

	delete(h.clientsByDevice, deviceID)
	delete(h.deviceByChannel, client.Channel)
	close(client.Channel)
}

func (h *Hub) SendToUser(userID uuid.UUID, msg BroadcastMessage) {
	for _, client := range h.clientsByUser[userID] {
		select {
		case client.Channel <- msg:
		default:
			h.cleanupDeadConnection(client.DeviceID)
		}
	}
}

func (h *Hub) SendToDevice(deviceID uuid.UUID, msg BroadcastMessage) {
	if client, ok := h.clientsByDevice[deviceID]; ok {
		select {
		case client.Channel <- msg:
		default:
			h.cleanupDeadConnection(deviceID)
		}
	}
}

func (h *Hub) cleanupDeadConnection(deviceID uuid.UUID) {
	// Delegate to unregisterDevice so cleanup is always complete.
	h.unregisterDevice(deviceID)
}

func (h *Hub) gracefulShutdown() {
	fmt.Println("Shutdown Hub Initiated")
	for deviceID, client := range h.clientsByDevice {
		delete(h.clientsByDevice, deviceID)
		delete(h.deviceByChannel, client.Channel)
		close(client.Channel)
	}
	h.clientsByUser = make(map[uuid.UUID][]*Client)
	h.usersByTopic = make(map[string]map[uuid.UUID]struct{})
	h.lastBroadCastedMessage = make(map[string]map[uuid.UUID]string)
	close(h.Done)
}

func (h *Hub) Shutdown() {
	close(h.shutdown)
	<-h.Done
	fmt.Println("Shutdown Hub Complete")
}
