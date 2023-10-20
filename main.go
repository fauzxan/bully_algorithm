package main

import (
	"core/client"
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	// "strings"
	"sync"
	"time"
)

func main() {
	me := client.Client{
		Lock: sync.Mutex{},
		// write code to set id
		// as soon as client joins, it needs to obtain replica from some arbitrary client in clientlist
		// the first client in the clientlist must always take the first value in the clientlist.json file
		// Subsequent clients may take on other values.
	}
	data, err := os.ReadFile("clientlist.json")
	if err != nil {
		fmt.Println("Error reading file clientlist.json")
	}
	var clientList map[int]string
	err = json.Unmarshal(data, &clientList)
	if err != nil {
		panic(err)
	}
	fmt.Println("Clientlist obtained!")
	me.Clientlist = clientList

	if len(me.Clientlist) == 0 {
		// If I am the only client so far, then I will set my self as both the coordinator, as well as communicate using the port
		me.Coordinator_id = 0
		me.Id = 0
		me.Clientlist[me.Id] = "127.0.0.1:3000"
	} else {
		// Otherwise, I will create a new clientid, and also a new port from which I will communicate with
		me.Id = me.GetUniqueRandom()
		me.Clientlist[me.Id] = "127.0.0.1:" + strconv.Itoa(3000+me.Id)
	}

	// Write back to json file
	jsonData, err := json.Marshal(me.Clientlist)
	if err != nil {
		fmt.Println("Could not marshal into json data")
		panic(err)
	}
	err = os.WriteFile("clientlist.json", jsonData, os.ModePerm)
	if err != nil {
		fmt.Println("Could not write into file")
		panic(err)
	}

	// Resolve and listen to address of self
	address, err := net.ResolveTCPAddr("tcp", me.Clientlist[me.Id])
	if err != nil {
		fmt.Println("Error resolving TCP address")
	}
	inbound, err := net.ListenTCP("tcp", address)
	if err != nil {
		fmt.Println("Could not listen to TCP address")

	}
	rpc.Register(&me)
	fmt.Println("Client is runnning at IP address", address)
	go rpc.Accept(inbound)

	go me.Sendcandidacy()

	random := ""
	for {
		fmt.Printf("Press enter for %d to communicate with coordinator.\n", me.Coordinator_id)
		fmt.Scanf("%s", &random)
		go me.Coordinatorsync()
		time.Sleep(1 * time.Second)
		fmt.Println("")
	}
}
