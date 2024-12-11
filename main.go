package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"sync"
)

type Message struct {
	ID    uint32  // 4 bytes
	Value float64 // 8 bytes
	Flag  byte    // 1 byte
}

func main() {
	// Channel to share data between threads
	messageChan := make(chan Message, 10) // Buffered channel to hold up to 10 messages

	var wg sync.WaitGroup

	wg.Add(1)
	go startServer(messageChan, &wg)

	// Start the logger
	wg.Add(1)
	go startLogger(messageChan, &wg)

	// Wait for both threads to finish (Ctrl+C will usually interrupt)
	wg.Wait()
}

// Function to handle a single connection
func startLogger(messageChan chan Message, wg *sync.WaitGroup) {
	defer wg.Done()

	for msg := range messageChan {
		// Print the recieved message
		fmt.Printf("Recieved Message: ID=%d, Value=%.2f, Flag=%c\n", msg.ID, msg.Value, msg.Flag)
	}
}

// TCP Server Function
func startServer(messageChan chan Message, wg *sync.WaitGroup) {
	defer wg.Done()

	// Listen on a TCP port
	listener, err := net.Listen("tcp", ":12345")
	if err != nil {
		fmt.Println("Error starting server: ", err)
		return
	}
	defer listener.Close()
	fmt.Println("Server is listening on port 12345...")

	for {
		// Accept a connection
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// Handle the connection in a separate goroutine
		go handleConnection(conn, messageChan)

	}

}

// Function to handle a single connection
func handleConnection(conn net.Conn, messageChan chan Message) {
	defer conn.Close() // Ensure the connection is closed when we're done

	buffer := make([]byte, 0, 1024) // Persistent buffer to store leftover data
	messageSize := 13               // Size of one message in bytes

	for {
		tempBuffer := make([]byte, 1024)

		// Read data into the temporary buffer
		n, err := conn.Read(tempBuffer)
		if err != nil {
			if err.Error() == "EOF" {
				fmt.Println("Connection closed by client.")
				return
			}
			fmt.Println("Error reading from connection:", err)
			return
		}

		buffer = append(buffer, tempBuffer[:n]...)

		for len(buffer) >= messageSize {
			// Extract one message
			messageData := buffer[:messageSize]

			// Remove the processed message from the buffer
			buffer = buffer[messageSize:]

			// Debug: Print raw data for the message
			fmt.Printf("Raw data recieved: %x\n", messageData)

			// Decode the binary data into a Message struct
			msg, err := decodeMessage(messageData)
			if err != nil {
				fmt.Println("Error decoding message: ", err)
				continue
			}
			// Send the message to the shared channel
			messageChan <- msg
		}

		if len(buffer) > 0 {
			fmt.Printf("Remaining buffer data: %x\n", buffer)
		}
	}

}

// Decode the binary data into a Message struct
func decodeMessage(data []byte) (Message, error) {
	var msg Message
	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.LittleEndian, &msg.ID) // Decode uint32
	if err != nil {
		return msg, fmt.Errorf("failed to decode ID: %v", err)
	}

	err = binary.Read(reader, binary.LittleEndian, &msg.Value) // Decode float64
	if err != nil {
		return msg, fmt.Errorf("failed to decode Value: %v", err)
	}

	err = binary.Read(reader, binary.LittleEndian, &msg.Flag) // Decode byte
	if err != nil {
		return msg, fmt.Errorf("failed to decode Flag: %v", err)
	}

	return msg, nil
}
