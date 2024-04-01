package main

import (
	"fmt"
	"ingresos_gastos/config"
	"ingresos_gastos/db"
	"ingresos_gastos/speaking"
	telegram "ingresos_gastos/telegram_bot_adapter"
	"log"
	"net/http"
)

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatalf("Failed to say OK: %v", err)
	}
}

// t@Gastos_Ingresos_bot
func main() {
	http.HandleFunc("/health", healthCheckHandler)

	// Start the HTTP server on port 8080
	go func() {
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	cfg := config.GetConfigFromEnv()
	fmt.Println(cfg)
	// Initialize the database
	storage := db.NewPostgresAdapter(cfg)
	bot, err := telegram.NewBotAdapter(cfg, storage)
	if err != nil {
		log.Fatalf("Failed to init Telegram Bot: %v", err)
	}

	env := speaking.MessagingPlatform{Storage: storage, Bot: bot}
	env.ListenToCommands()
	env.ListenToUserInput()
	env.ListenToInlineActions()
}
