package adapter

import (
    "encoding/json"
    "log"
    "net/http"
    "WS/websoquet-server/internal/domain"
    "WS/websoquet-server/internal/service"

    "github.com/gorilla/websocket"
)

// Handler gestiona conexiones WebSocket.
type Handler struct {
    Service *service.WebsoquetService
}

func NewHandler(svc *service.WebsoquetService) *Handler {
    return &Handler{Service: svc}
}

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool { return true },
}

// ServeWS convierte HTTP en WebSocket y registra al cliente.
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Println("Upgrade error:", err)
        return
    }
    client := NewClient(conn)

    // Leer accountID de la query: ?account=Colmena1
    accountID := r.URL.Query().Get("account")
    if accountID == "" {
        log.Println("Missing accountID")
        client.Close()
        return
    }

    h.Service.RegisterClient(accountID, client)
    go h.handleMessages(accountID, client)
}

func (h *Handler) handleMessages(accountID string, client *Client) {
    defer h.Service.RemoveClient(accountID, client)
    for {
        _, msg, err := client.ReadMessage()
        if err != nil {
            log.Println("Read error:", err)
            break
        }
        var m domain.Message
        if err := json.Unmarshal(msg, &m); err != nil {
            log.Println("Invalid JSON:", err)
            continue
        }
        log.Printf("Received from %s → %s", m.Sender, m.Receiver)
        // Reenvía el mensaje según la lógica existente
        msgOut, _ := json.Marshal(m)
        h.Service.SendMessageToAccount(m.Receiver, msgOut)
    }
}
