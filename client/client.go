package client

import (
	"fmt"
	"net/rpc"
	"sync"
)


type Client struct{
	Lock sync.Mutex
	Id int
	Coordinator_id int
	Clientlist map[int]string
	Replica []int
}

// Flag variables:
var Election_invoked = false
var Higherid = false

var ANNOUNCE = "announce"
var SYNC = "sync"
var ACK = "ack"
var VICTORY = "victory"

func (client *Client) HandleCommunication(message *Message, reply *Message) error{ // Handle communication
	client.Lock.Lock()
	defer client.Lock.Unlock()

	fmt.Println("Handling Message type:", message.Type)
	if (message.Type == SYNC){
		fmt.Println("Received a sync request from", message.From)
		reply.Replica = client.Replica
	}else if(message.Type == ANNOUNCE){
		fmt.Println("Received an announcement from", message.From)
		reply.Type = "ack"
		go client.Sendcandidacy()
	}else if(message.Type == VICTORY){
		fmt.Println("Received a victory message from", message.From)
		client.Coordinator_id = message.From
		client.Clientlist = message.Clientlist
		fmt.Println(client.Coordinator_id, "is now my coordinator")
		reply.Type = "ack"
	}
	return nil
}

func (client *Client) Coordinatorsync(){ // communicatetocoordinator
	if (client.Id == client.Coordinator_id){return}
	coordinator_ip := client.Clientlist[client.Coordinator_id]
	fmt.Println("Getting replica from coordinator", client.Coordinator_id)
	reply := Message{}
	send := Message{Type: SYNC, From: client.Id}
	clnt, err := rpc.Dial("tcp", coordinator_ip)
	if (err != nil){
		fmt.Println("There was an error trying to connect to the coordinator")
		fmt.Println(err)
		fmt.Println("Invoking election")
		// code to invoke election here
		go client.Sendcandidacy()
		return
	}

	err = clnt.Call("Client.HandleCommunication", send, &reply)
	if (err != nil){
		fmt.Println("There was an error trying to connect to the coordinator")
		fmt.Println(err)
		fmt.Println("Invoking election")
		// code to invoke election here
		go client.Sendcandidacy()
	}
	client.Replica = reply.Replica
	fmt.Printf("Replica has been updated:%v", client.Replica)
}


func (client *Client) Sendcandidacy(){ // invokeelection()
	fmt.Println("Discovery phase beginning...")
	for id, ip := range client.Clientlist{
		reply := Message{}
		send := Message{Type: ANNOUNCE, From: client.Id}
		if id > client.Id{
			fmt.Println("Sending candidacy to", id)
			clnt, err := rpc.Dial("tcp", ip)
			if (err != nil){
				fmt.Println("Communication to", id, "failed.")
				continue
			}
			err = clnt.Call("Client.HandleCommunication", send, &reply)
			if (err != nil){
				fmt.Println("Communication to", id, "failed")
				continue
			}
			if (reply.Type == "ack"){
				client.Lock.Lock()
				fmt.Println("Received an ACK reply from", id)
				Higherid = true
				client.Lock.Unlock()
			}
		}
	}
	if (Higherid == false){
		// function to make yourself coordinator
		client.Coordinator_id = client.Id
		go client.Announcevictory()
	}
	Election_invoked = true
}

func (client *Client) Announcevictory(){ // Make yourself coordinator
	// client.Lock.Lock()
	// defer client.Lock.Unlock()
	send := Message{Type: VICTORY, From: client.Id, Clientlist: client.Clientlist}
	reply := Message{}
	fmt.Println("No higher id node found. I am announcing victory!")
	client.Printclients()
	for id, ip := range client.Clientlist{
		clnt, err := rpc.Dial("tcp", ip)
		if (err != nil){
			fmt.Println("Communication to", id, "failed")
			continue
		}
		err = clnt.Call("Client.HandleCommunication", send, &reply)
		if (reply.Type == "ack" && id > client.Id){
			// if you are announcing vicotry and you receive an ack message from a higher id, then just stop announcing
			fmt.Println("Message sent to", id, "successfully")
			fmt.Println("Client", id, "is awake")
			break
		}else if(reply.Type == "ack"){
			fmt.Println("Client", id, "acknowledged me as coordinator")
		}
	}
	return
}

/*
***************************
UTILITY FUNCTIONS
***************************
*/

func (client *Client) Printclients(){
	for id, ip := range client.Clientlist{
		fmt.Println(id, ip)
	}
}

func (client *Client) Getmax() int{
	client.Lock.Lock()
	defer client.Lock.Unlock()
	max := 0
	for id, _ := range client.Clientlist{
		if (id > max){max = id}
	}
	return max
}

func (client *Client) Check(n int) bool{
	client.Lock.Lock()
	defer client.Lock.Unlock()
	for id := range client.Clientlist {
		if (n == id){
			return false
		}
	}
	return true
}

func (client *Client) GetUniqueRandom() int{
	for i:=0;;i++{
		check := client.Check(i)
		if (check == false){return i}
	}
}