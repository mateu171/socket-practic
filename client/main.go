package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println("Помилка підключення:", err)
		return
	}

	defer conn.Close()

	fmt.Println("Підключено до сервера.")
	fmt.Println("Введіть команду:")

	go func() {
		reader := bufio.NewReader(conn)

		for {
			message, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			fmt.Print(message)
		}
	}()

	input := bufio.NewScanner(os.Stdin)

	for input.Scan() {
		message := input.Text() + "\n"

		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Помилка відправлення:", err)
			return
		}
	}
}
