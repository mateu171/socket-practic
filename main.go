package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	clients map[net.Conn]bool
	mu      sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients: make(map[net.Conn]bool),
	}
}

func (s *Server) addClient(conn net.Conn) {
	s.mu.Lock()
	s.clients[conn] = true
	s.mu.Unlock()

	log.Printf("[ПІДКЛЮЧЕННЯ] Новий клієнт підключився: %s", conn.RemoteAddr())

	s.broadcastSystem(
		fmt.Sprintf("[СИСТЕМА] Підключився новий клієнт: %s", conn.RemoteAddr()),
		conn,
	)
}

func (s *Server) removeClient(conn net.Conn) {
	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()

	log.Printf("[ВІДКЛЮЧЕННЯ] Клієнт відключився: %s", conn.RemoteAddr())

	s.broadcastSystem(
		fmt.Sprintf("[СИСТЕМА] Клієнт вийшов: %s", conn.RemoteAddr()),
		conn,
	)
}

func (s *Server) broadcast(message string, sender net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if client != sender {
			_, err := client.Write([]byte(message + "\n"))
			if err != nil {
				log.Printf("[ПОМИЛКА] Не вдалося надіслати повідомлення: %v", err)
			}
		}
	}
}

func (s *Server) broadcastSystem(message string, excluded net.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for client := range s.clients {
		if client != excluded {
			_, err := client.Write([]byte(message + "\n"))
			if err != nil {
				log.Printf("[ПОМИЛКА] Не вдалося надіслати системне повідомлення: %v", err)
			}
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	defer s.removeClient(conn)

	_, err := conn.Write([]byte(
		"[СЕРВЕР] Ласкаво просимо до мережевого сервісу!\n",
	))

	if err != nil {
		return
	}

	reader := bufio.NewReader(conn)

	for {
		message, err := reader.ReadString('\n')

		if err != nil {
			return
		}

		message = strings.TrimSpace(message)

		if message == "" {
			continue
		}

		parts := strings.SplitN(message, " ", 2)
		command := strings.ToUpper(parts[0])

		switch command {
		case "ECHO":
			if len(parts) < 2 {
				conn.Write([]byte("[ПОМИЛКА]: використання ECHO <повідомлення>\n"))
				continue
			}

			response := fmt.Sprintf("[ECHO]: %s\n", parts[1])
			conn.Write([]byte(response))

		case "TIME":
			response := fmt.Sprintf(
				"[TIME]: %s\n",
				time.Now().Format(time.RFC3339),
			)

			conn.Write([]byte(response))

		case "BROADCAST", "MSG":
			if len(parts) < 2 {
				conn.Write([]byte(
					"[ПОМИЛКА]: використання BROADCAST <повідомлення>\n",
				))
				continue
			}

			text := fmt.Sprintf(
				"[СПОВІЩЕННЯ від %s]: %s",
				conn.RemoteAddr(),
				parts[1],
			)

			log.Printf(
				"[РОЗСИЛКА] Від %s -> \"%s\"",
				conn.RemoteAddr(),
				parts[1],
			)

			s.broadcast(text, conn)

		case "CLOSE", "QUIT":
			conn.Write([]byte("[СЕРВЕР] До побачення!\n"))
			return

		default:
			conn.Write([]byte(
				"[ПОМИЛКА]: невідома команда. Доступні команди: ECHO, TIME, BROADCAST, MSG, CLOSE, QUIT\n",
			))
		}
	}
}

func main() {
	server := NewServer()

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatal(err)
	}

	defer listener.Close()

	fmt.Println("=== TCP-сервер диспетчеризації ===")
	fmt.Println("[СЕРВЕР] TCP-слухач успішно відкрито на порту 9000")
	fmt.Println("[СЕРВЕР] Очікування підключень...")

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Printf("[ПОМИЛКА] Accept: %v", err)
			continue
		}

		server.addClient(conn)

		go server.handleConnection(conn)
	}
}