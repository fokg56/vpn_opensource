package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/vpn-client/internal/web"
)

func main() {
	// Флаги командной строки
	webAddr := flag.String("addr", "0.0.0.0:8080", "Адрес для веб-сервера (например, 0.0.0.0:8080)")
	flag.Parse()

	// Создаем и запускаем веб-сервер
	server := web.NewServer(*webAddr)

	// Запускаем сервер в отдельной goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Ошибка запуска веб-сервера: %v", err)
		}
	}()

	// Обработка сигналов для корректного завершения
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	fmt.Println("\nПолучен сигнал завершения, выключаемся...")
	os.Exit(0)
}
