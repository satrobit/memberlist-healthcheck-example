package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/hashicorp/memberlist"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type (
	Node struct {
		memberlist *memberlist.Memberlist
	}
)

type Item struct {
	Ip     string `json:"ip"`
	Status string `json:"status"`
}

func initCluster(bindIP, httpPort string) {

	clusterKey := make([]byte, 32)
	_, err := rand.Read(clusterKey)
	if err != nil {
		panic(err)
	}

	config := memberlist.DefaultLocalConfig()
	config.BindAddr = bindIP
	config.Name = bindIP
	config.SecretKey = clusterKey

	ml, err := memberlist.Create(config)

	if err != nil {
		panic(err)
	}

	node := Node{
		memberlist: ml,
	}

	log.Printf("new cluster created. key: %s\n", base64.StdEncoding.EncodeToString(clusterKey))

	http.HandleFunc("/", node.handler)

	go func() {
		http.ListenAndServe(":"+httpPort, nil)
	}()

	log.Printf("webserver is up. URL: http://%s:%s/ \n", bindIP, httpPort)

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := ml.Leave(time.Second * 5); err != nil {
			panic(err)
		}

	}
}

func joinCluster(bindIP, httpPort, clusterKey, knownIP string) {
	config := memberlist.DefaultLocalConfig()
	config.BindAddr = bindIP
	config.Name = bindIP
	config.SecretKey, _ = base64.StdEncoding.DecodeString(clusterKey)

	ml, err := memberlist.Create(config)

	if err != nil {
		panic(err)
	}

	node := Node{
		memberlist: ml,
	}

	_, err = ml.Join([]string{knownIP})
	if err != nil {
		panic("Failed to join cluster: " + err.Error())
	}

	log.Printf("Joined the cluster")

	http.HandleFunc("/", node.handler)

	go func() {
		http.ListenAndServe(":"+httpPort, nil)
	}()

	log.Printf("webserver is up. URL: http://%s:%s/ \n", bindIP, httpPort)

	incomingSigs := make(chan os.Signal, 1)
	signal.Notify(incomingSigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, os.Interrupt)

	select {
	case <-incomingSigs:
		if err := ml.Leave(time.Second * 5); err != nil {
			panic(err)
		}

	}

}

func main() {

	joinCmd := flag.NewFlagSet("join", flag.ExitOnError)
	joinClusterKey := joinCmd.String("cluster-key", "", "cluster-key")
	joinKnownIP := joinCmd.String("known-ip", "", "known-ip")
	joinBindIP := joinCmd.String("bind-ip", "127.0.0.1", "bind-ip")
	joinHttpPort := joinCmd.String("http-port", "8888", "http-port")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initBindIP := initCmd.String("bind-ip", "127.0.0.1", "bind-ip")
	initHttpPort := initCmd.String("http-port", "8888", "http-port")

	if len(os.Args) < 2 {
		fmt.Println("expected 'join' or 'init' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {

	case "join":
		joinCmd.Parse(os.Args[2:])
		joinCluster(*joinBindIP, *joinHttpPort, *joinClusterKey, *joinKnownIP)
	case "init":
		initCmd.Parse(os.Args[2:])
		initCluster(*initBindIP, *initHttpPort)
	default:
		fmt.Println("expected 'join' or 'init' subcommands")
		os.Exit(1)
	}

	os.Exit(0)

}

func (n *Node) handler(w http.ResponseWriter, req *http.Request) {

	var items []Item

	for _, member := range n.memberlist.Members() {
		hostName := member.Addr.String()
		portNum := "80"
		seconds := 5
		timeOut := time.Duration(seconds) * time.Second
		conn, err := net.DialTimeout("tcp", hostName+":"+portNum, timeOut)

		if err != nil {
			items = append(items, Item{Ip: conn.RemoteAddr().String(), Status: "DOWN"})
		} else {
			items = append(items, Item{Ip: conn.RemoteAddr().String(), Status: "UP"})
		}
	}

	js, err := json.Marshal(items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)

}
