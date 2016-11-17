package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func init() {
	go StartServer(8000)
	time.Sleep(time.Second)
}

func create(t *testing.T) string {
	response, err := http.Get("http://127.0.0.1:8000/create")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("status != 200")
	}
	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	containerID := string(bs)
	if len(containerID) != 64 {
		t.Fatal("bad container ID")
	}
	return containerID
}

func TestCreate(t *testing.T) {
	create(t)
}

func TestCommand(t *testing.T) {
	containerID := create(t)
	response, err := http.Get(fmt.Sprintf("http://127.0.0.1:8000/command/%s?command=echo+hello+world", containerID))
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatal("status != 200")
	}
	bs, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(bs) != "hello world\n" {
		t.Fatal("expected hello world")
	}
}

func runCommand(command string, connection *websocket.Conn, t *testing.T) string {
	if err := connection.WriteMessage(websocket.TextMessage, []byte(command)); err != nil {
		t.Fatal(err)
	}

	messageType, output, err := connection.ReadMessage()
	if messageType != websocket.TextMessage {
		t.Fatal("websocket: messageType != TextMessage")
	}
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func TestSession(t *testing.T) {
	containerID := create(t)

	connection, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:8000/session/%s", containerID), nil)
	if err != nil {
		t.Fatal(err)
	}

	runCommand("PS1=''", connection, t)     // disable command prompt
	runCommand("stty -echo", connection, t) // disable input echo
	time.Sleep(time.Second)                 // TODO(kvu787): delay for stty -echo to take effect, why?

	commands := []string{"echo hello world", "badcommand", "cat > hello.txt"}
	wantedOutput := []string{"hello world", "bash: badcommand: command not found", ""}
	for i, command := range commands {
		got := runCommand(command, connection, t)
		want := wantedOutput[i]
		if got != want {
			t.Fatalf("got %v, want %s", got, want)
		}
	}
}
