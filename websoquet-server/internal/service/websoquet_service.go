package service

import (
	"encoding/json"
	"log"
	"sync"

	"WS/websoquet-server/internal/domain"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	mqttBroker = "tcp://13.223.36.70:1883"
	mqttTopic  = "colmena/data"
)

type WebsoquetService struct {
	Clients map[string][]domain.Client
	mu      sync.Mutex
	mqttCli mqtt.Client
}

func NewWebsoquetService() *WebsoquetService {
	s := &WebsoquetService{
		Clients: make(map[string][]domain.Client),
	}
	s.initMQTT()
	return s
}


func (s *WebsoquetService) initMQTT() {
	opts := mqtt.NewClientOptions().AddBroker(mqttBroker)
	opts.SetClientID("ws-subscriber")
	opts.SetAutoReconnect(true)

	s.mqttCli = mqtt.NewClient(opts)
	if tok := s.mqttCli.Connect(); tok.Wait() && tok.Error() != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", tok.Error())
	}
	log.Printf("Connected to MQTT broker %s", mqttBroker)

	if tok := s.mqttCli.Subscribe(mqttTopic, 0, s.onMQTTMessage); tok.Wait() && tok.Error() != nil {
		log.Fatalf("Failed to subscribe to %s: %v", mqttTopic, tok.Error())
	}
	log.Printf("Subscribed to MQTT topic %s", mqttTopic)
}

// onMQTTMessage se llama al recibir un mensaje MQTT.
func (s *WebsoquetService) onMQTTMessage(_ mqtt.Client, msg mqtt.Message) {
	var m domain.Message
	if err := json.Unmarshal(msg.Payload(), &m); err != nil {
		log.Printf("Invalid MQTT payload: %v", err)
		return
	}
	s.SendMessageToAccount(m.Receiver, msg.Payload())
}

func (s *WebsoquetService) RegisterClient(accountID string, client domain.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Clients[accountID] = append(s.Clients[accountID], client)
	log.Printf("Client registered: %s", accountID)
}

// SendMessageToAccount envía el mensaje a todos los clientes de ese receiver.
func (s *WebsoquetService) SendMessageToAccount(receiver string, msg []byte) {
	s.mu.Lock()
	clients := s.Clients[receiver]
	s.mu.Unlock()

	if len(clients) == 0 {
		log.Printf("No clients for receiver %s", receiver)
		return
	}
	for _, c := range clients {
		if err := c.WriteMessage(1 /* TextMessage */, msg); err != nil {
			log.Printf("Error sending to %s: %v", receiver, err)
			s.RemoveClient(receiver, c)
		}
	}
}

// RemoveClient elimina un cliente desconectado.
func (s *WebsoquetService) RemoveClient(accountID string, toRemove domain.Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := s.Clients[accountID]
	newList := make([]domain.Client, 0, len(list))
	for _, c := range list {
		if c != toRemove {
			newList = append(newList, c)
		} else {
			c.Close()
			log.Printf("Client disconnected: %s", accountID)
		}
	}
	if len(newList) == 0 {
		delete(s.Clients, accountID)
	} else {
		s.Clients[accountID] = newList
	}
}
