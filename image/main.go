package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/kr/pty"
)

var commandShell = []string{"/bin/bash", "-c"}
var sessionShell = []string{"/bin/bash", "--login"}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	StartServer(10411)
}

func StartServer(port int) {
	r := mux.NewRouter()
	r.HandleFunc("/command", ShellCommandHandler)
	r.HandleFunc("/session", SessionHandler)
	r.HandleFunc("/xterm", XtermHandler)
	log.Printf("Starting server on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
	time.Sleep(time.Second)
}

func ShellCommandHandler(response http.ResponseWriter, request *http.Request) {
	commands, found := request.URL.Query()["command"]
	if !found {
		response.WriteHeader(http.StatusBadRequest)
		response.Write([]byte("command field in query parameter is missing"))
		return
	}

	for _, command := range commands {
		output, err := exec.Command(commandShell[0], append(commandShell[1:], command)...).Output()
		if err != nil {
			log.Fatal(err)
		}
		response.Write(output)
	}
}

func XtermHandler(response http.ResponseWriter, request *http.Request) {
	file, err := pty.Start(exec.Command(sessionShell[0], sessionShell[1:]...))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Fatal(err)
	}

	// wait for terminal session to start
	// TODO(kvu787): find a better way
	time.Sleep(250 * time.Millisecond)

	// ws -> shell
	go func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if messageType != websocket.TextMessage {
				log.Fatal("websocket: messageType != TextMessage")
			}
			if err != nil {
				log.Println(err)
				break
			}
				io.Copy(file, bytes.NewReader(data))
		}
	}()

	// shell -> ws
	go func() {
		reader := bufio.NewReader(file)
		for {
			b, err := reader.ReadByte()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatal(err)
				}
			}
			err = conn.WriteMessage(websocket.TextMessage, []byte{b})
			if err != nil {
				log.Println(err)
				break
			}
		}
	}()
}

func SessionHandler(response http.ResponseWriter, request *http.Request)  {
	file, err := pty.Start(exec.Command(sessionShell[0], sessionShell[1:]...))
	if err != nil {
		log.Fatal(err)
	}

	conn, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Fatal(err)
	}

	// wait for terminal session to start
	// TODO(kvu787): find a better way
	time.Sleep(250 * time.Millisecond)

	// ws -> shell
	go func() {
		for {
			messageType, data, err := conn.ReadMessage()
			if messageType != websocket.TextMessage {
				log.Fatal("websocket: messageType != TextMessage")
			}
			if err != nil {
				log.Println(err)
				break
			}
				io.Copy(file, bytes.NewReader(append(data, '\n')))
		}
	}()

	// shell -> ws
	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			output := scanner.Text()
			err := conn.WriteMessage(websocket.TextMessage, []byte(output))
			if err != nil {
				log.Println(err)
				break
			}
		}
		err := scanner.Err()
		if err != nil {
			log.Fatal(err)
			return
		}
	}()
}
