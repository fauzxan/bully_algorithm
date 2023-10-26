package main

import (
	"core/client"
	"encoding/json"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
	"github.com/fatih/color"
)

var menu = color.New(color.FgCyan).Add(color.BgBlack)

func main() {
	me := client.Client{Lock: sync.Mutex{}}

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

	var input string
	for {
		// menu.Printf("Press any button for %d to communicate with coordinator.\n", me.Coordinator_id)
		menu.Println("Press 1 to enter something into my replica")
		menu.Println("Press 2 to view my replica")
		menu.Println("Press 3 to view my clientlist")
		fmt.Scanln(&input)
		if input == "1"{
			var repinput string
			menu.Println("Enter one number into the replica")
			fmt.Scanln(&repinput)
			num, err := strconv.Atoi(repinput)
			if err != nil{
				fmt.Println("Enter a number lah")
			}
			go me.EnterReplica(num)
		}else if (input == "2"){
			menu.Println("Replica", me.Replica)
		}else if (input == "3"){
			me.Printclients()
		}
		
		
		time.Sleep(200)
		menu.Println("")
	}
}
