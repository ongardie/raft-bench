package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/goraft/raft"
)

var (
	port       = flag.Int("p", 7001, "raft port")
	hostname   = flag.String("h", "localhost", "raft hostname")
	verbose    = flag.Bool("v", false, "debug info")
	leader     = flag.String("l", "", "the address of the leader")
	peerNumber = flag.Int("peer-number", 3, "number of expected peers")
)

var peerCnt = 1

func main() {
	runtime.GOMAXPROCS(2)
	flag.Parse()
	if *verbose == true {
		raft.SetLogLevel(2)
	}

	t := raft.NewHTTPTransporter("/raft")

	name := fmt.Sprintf("raft-%v", *port)
	dir := fmt.Sprintf("./raft-%v", *port)
	connStr := fmt.Sprintf("http://%v:%v", *hostname, *port)

	os.RemoveAll(dir)
	os.Mkdir(dir, 0700)

	server, err := raft.NewServer(name, dir, t, nil, nil, connStr)
	if err != nil {
		panic(err)
	}

	server.SetHeartbeatTimeout(time.Millisecond * 2)

	err = server.Start()
	if err != nil {
		panic(err)
	}

	// Create listener for HTTP server and start it.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	joinCommand := &raft.DefaultJoinCommand{
		Name:             name,
		ConnectionString: connStr,
	}

	if *leader != "" {
		var b bytes.Buffer
		json.NewEncoder(&b).Encode(joinCommand)
		resp, err := http.Post(fmt.Sprintf("http://%s/join", *leader), "application/json", &b)
		if err != nil {
			panic(err)
		}

		resp.Body.Close()
		fmt.Println("start as a follower of ", *leader)
	} else {
		_, err = server.Do(joinCommand)
		if err != nil {
			panic(err)
		}

		server.AddEventListener("addPeer", createAddListenerFunc(server))
		fmt.Println("start as a leader")
	}
	server.AddEventListener("stateChange", createStateChangeListenerFunc(server))
	// Create wrapping HTTP server.
	mux := http.NewServeMux()
	t.Install(server, mux)
	mux.HandleFunc("/join", createJoinHandler(server))
	httpServer := &http.Server{Addr: connStr, Handler: mux}
	httpServer.Serve(listener)
}

func createJoinHandler(s raft.Server) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		fmt.Println("recv join!")
		command := &raft.DefaultJoinCommand{}

		if err := json.NewDecoder(req.Body).Decode(&command); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if _, err := s.Do(command); err != nil {
			fmt.Println("error...", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func createAddListenerFunc(s raft.Server) func(e raft.Event) {
	return func(e raft.Event) {
		fmt.Println("add peer")
		if len(s.Peers()) == *peerNumber-1 {
			go send(s, 1000)
		}
	}
}

func createStateChangeListenerFunc(s raft.Server) func(e raft.Event) {
	return func(e raft.Event) {
		if e.Value() == "leader" {
			go send(s, 1000)
		}
	}
}
