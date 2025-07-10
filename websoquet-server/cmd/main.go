package main

import (
	"fmt"
	"log"
	"WS/websoquet-server/internal/app"
	"os"
	"github.com/joho/godotenv"
)

func main() {

	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error cargando archivo .env")
	}
	fmt.Println("Broker:", os.Getenv("MQTT_BROKER"))
	fmt.Println("Topic:", os.Getenv("MQTT_TOPIC"))


	server := app.NewServer()
	addr := ":8080"
	fmt.Printf("Servidor WebSocket ejecutándose en %s\n", addr)
	if err := server.Start(addr); err != nil {
		log.Fatal("Error en el servidor:", err)
	}
}
