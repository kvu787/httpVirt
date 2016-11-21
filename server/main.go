package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rs/cors"
)

var id2port = map[string]int{}
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	useCors := len(os.Args) >= 2 && os.Args[1] == "cors"
	StartServer(10411, useCors)
}

func StartServer(port int, useCors bool) {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	// Setup routing
	r := mux.NewRouter()
	r.HandleFunc("/create", CreateHandler).Methods("GET")
	r.HandleFunc("/command/{containerID}", ShellCommandHandler)
	r.HandleFunc("/session/{containerID}", ShellSessionHandler)
	r.HandleFunc("/xterm/{containerID}", ShellXtermHandler)

	// wrap router in cors handler
	// let other domains ping our service to make debugging easier
	var handler http.Handler = r
	if useCors {
		handler = cors.Default().Handler(r)
	}

	// Start server
	log.Printf("Starting server on port %d...\n", port)
	log.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), handler))
}

func CreateHandler(w http.ResponseWriter, r *http.Request) {
	// Run httpvirt image in a new container
	output, err := exec.Command("docker", "run", "-d", "-P", "httpvirt").Output()
	if err != nil {
		log.Println(err)
		return
	}
	containerID := strings.TrimSpace(string(output))

	// Retrieve the host port this container was mapped to
	output, err = exec.Command(
		"docker",
		"inspect",
		"--format='{{range .NetworkSettings.Ports}}{{with index . 0}}{{.HostPort}}{{end}}{{end}}'",
		containerID).Output()
	if err != nil {
		log.Println(err)
		return
	}
	port, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		log.Println(err)
		return
	}

	// Store (containerID, port)
	id2port[containerID] = port

	// reply with containerID
	w.Write([]byte(containerID))

	// wait for server inside container to start
	// TODO(kvu787): find a better way
	time.Sleep(time.Second)
}

func ShellCommandHandler(response http.ResponseWriter, request *http.Request) {
	containerID := mux.Vars(request)["containerID"]
	port := id2port[containerID]

	// Forward request to container
	containerResponse, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/command?%s", port, request.URL.Query().Encode()))
	if err != nil {
		log.Println(err)
		return
	}

	// Forward container response
	io.Copy(response, containerResponse.Body)
	containerResponse.Body.Close()
}

func forward(src, dst *websocket.Conn) {
	for {
		messageType, data, err := src.ReadMessage()
		if messageType != websocket.TextMessage {
			log.Println("websocket: messageType != TextMessage")
			return
		}
		if err != nil {
			log.Println(err)
			break
		}
		err = dst.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println(err)
			break
		}
	}
}

func ShellSessionHandler(response http.ResponseWriter, request *http.Request) {
	containerID := mux.Vars(request)["containerID"]
	port := id2port[containerID]

	url := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%d", port), Path: "/session"}
	containerConnection, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	userConnection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// ws -> shell
	go forward(containerConnection, userConnection)
	go forward(userConnection, containerConnection)
}

func ShellXtermHandler(response http.ResponseWriter, request *http.Request) {
	containerID := mux.Vars(request)["containerID"]
	port := id2port[containerID]

	url := url.URL{Scheme: "ws", Host: fmt.Sprintf("127.0.0.1:%d", port), Path: "/xterm"}
	containerConnection, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		log.Println(err)
		return
	}

	userConnection, err := upgrader.Upgrade(response, request, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// ws -> shell
	go forward(containerConnection, userConnection)
	go forward(userConnection, containerConnection)
}
