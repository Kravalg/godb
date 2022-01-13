package main

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

const (
	ConnHost = "localhost"
	ConnPort = "6666"
	ConnType = "tcp"
)

type Storage struct {
	rwm   sync.RWMutex
	cache map[string]string
}

func NewStorage() *Storage {
	return &Storage{
		cache: make(map[string]string),
	}
}

func (s *Storage) delete(key string) {
	delete(s.cache, s.clearStringFromNulBytes(key))
}

func (s *Storage) set(key string, value string) {
	s.cache[s.clearStringFromNulBytes(key)] = s.clearStringFromNulBytes(value)
}

func (s *Storage) get(key string) string {
	value, exists := s.cache[s.clearStringFromNulBytes(key)]
	if exists {
		return value
	}
	return "nil"
}

func (s *Storage) clearStringFromNulBytes(key string) string {
	return string(bytes.Trim([]byte(key), "\x00"))
}

func (s *Storage) Delete(key string) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.delete(key)
}

func (s *Storage) Set(key string, value string) {
	s.rwm.Lock()
	defer s.rwm.Unlock()
	s.set(key, value)
}

func (s *Storage) Get(key string) string {
	s.rwm.RLock()
	defer s.rwm.RUnlock()
	return s.get(key)
}

type Action struct {
	method string
	key    string
	value  string
	done   bool
}

func main() {
	storage := NewStorage()
	l, err := net.Listen(ConnType, ConnHost+":"+ConnPort)
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}

	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			fmt.Println("Close the listener when the application closes", err.Error())
		}
	}(l)

	fmt.Println("Listening on " + ConnHost + ":" + ConnPort)

	data := make(chan Action)

	go func() {
		for action := range data {
			if action.done {
				continue
			}

			switch action.method {
			case "GET":
				action.done = true
				action.value = storage.Get(action.key)
				data <- action
			case "SET":
				action.done = true
				storage.Set(action.key, action.value)
				data <- action
			case "DELETE":
				action.done = true
				storage.Delete(action.key)
				data <- action
			}
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}

		buf := make([]byte, 1024)
		_, readErr := conn.Read(buf)
		if readErr != nil {
			fmt.Println("Error reading:", readErr.Error())
		}

		commands := strings.Split(string(buf), ";")

		go handleRequest(commands, conn, data)
	}
}

func handleRequest(commands []string, conn net.Conn, data chan Action) {
	for key := range commands {
		command := commands[key]

		commandParts := strings.Split(command, " ")

		if len(commandParts) < 2 {
			fmt.Println("Unknown command")
		}

		switch commandParts[0] {
		case "GET":
			handleGetCommand(conn, data, commandParts[1])
		case "SET":
			handleSetCommand(conn, data, commandParts[1], commandParts[2])
		case "DELETE":
			handleDeleteCommand(conn, data, commandParts[1])
		default:
			_, _ = conn.Write([]byte("Unknown command"))
		}
		_, _ = conn.Write([]byte("\n"))
	}
}

func handleGetCommand(conn net.Conn, data chan Action, key string) {
	data <- Action{method: "GET", key: key, value: "", done: false}
	value := <-data

	if value.done {
		_, _ = conn.Write([]byte(value.value))
	}
}

func handleSetCommand(conn net.Conn, data chan Action, key string, value string) {
	data <- Action{method: "SET", key: key, value: value, done: false}
	newValue := <-data

	if newValue.done {
		_, _ = conn.Write([]byte("OK"))
	}
}

func handleDeleteCommand(conn net.Conn, data chan Action, key string) {
	data <- Action{method: "DELETE", key: key, value: "", done: false}
	newValue := <-data

	if newValue.done {
		_, _ = conn.Write([]byte("OK"))
	}
}
