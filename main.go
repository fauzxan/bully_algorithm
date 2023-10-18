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
)

func main(){
	me := client.Client{
		Lock: sync.Mutex{},
		// write code to set id
		// as soon as client joins, it needs to obtain replica from some arbitrary client in clientlist
		// the first client in the clientlist must always take the first value in the clientlist.json file
		// Subsequent clients may take on other values.
	}
	data, err := os.ReadFile("clientlist.json")
	if (err != nil){
		fmt.Println("Error reading file clientlist.json")
	}
	var clientList map[int]string
    err = json.Unmarshal(data, &clientList)
    if err != nil {
        panic(err)
    }
	fmt.Println("Clientlist obtained!")
	me.Clientlist = clientList
	for clientID, clientAddress := range clientList {
        println(clientID, clientAddress)
    }



	if (len(me.Clientlist) == 1){
		// If I am the only client so far, then I will set my self as both the coordinator, as well as communicate using the port
		me.Coordinator_id = 0
		me.Id = 0
		
	}else{
		// Otherwise, I will create a new clientid, and also a new port from which I will communicate with
		me.Id = len(clientList)
		me.Clientlist[me.Id] = "127.0.0.1:300" + strconv.Itoa(me.Id)
		jsonData, err := json.Marshal(clientList)
		if err != nil {
			fmt.Println("Could not marshal into json data")
			panic(err)
		}

		// Write the JSON data to the file.
		err = os.WriteFile("clientlist.json", jsonData, os.ModePerm)
		if err != nil {
			fmt.Println("Could not write into file")
			panic(err)
    	}
	}
	address, err := net.ResolveTCPAddr("tcp", me.Clientlist[me.Id]) 
	if err != nil{
		fmt.Println("Error resolving TCP address")
	}
	inbound, err := net.ListenTCP("tcp", address)
	if err != nil{
		fmt.Println("Could not listen to TCP address")
		
	}
	rpc.Register(&me)
	fmt.Println("Client is runnning at IP address", address)
	rpc.Accept(inbound)

	go me.Sendcandidacy()
	time.Sleep(2*time.Second)

	random := ""
	for{
		fmt.Printf("Press enter for %d to communicate with coordinator.\n", me.Coordinator_id)
		fmt.Scanf("%s", &random)
		go me.Coordinatorsync()
		time.Sleep(1* time.Second)
		fmt.Println("")
	}
}