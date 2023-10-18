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


func (client *Client) HandleCommunication(client_id int, reply *Message) error{ // Handle communication
	client.Lock.Lock()
	defer client.Lock.Unlock()

	fmt.Println("Received a sync request from", client_id)
	if (reply.Type == "sync"){
		reply.Replica = client.Replica
	}else if(reply.Type == "announce"){
		reply.Type = "ack"
		client.Sendcandidacy()
	}else if(reply.Type == "victory"){
		client.Coordinator_id = client_id
		reply.Type = "ack"
	}
	return nil
}

func (client *Client) Coordinatorsync(){ // communicatetocoordinator
	coordinator_ip := client.Clientlist[client.Coordinator_id]
	fmt.Println("Getting replica from coordinator", client.Coordinator_id)
	reply := Message{
		Type: "sync",
	}
	clnt, err := rpc.Dial("tcp", coordinator_ip)
	if (err != nil){
		fmt.Println("There was an error trying to connect to the coordinator")
		fmt.Println("Invoking election")
		// code to invoke election here
		return
	}
	err = clnt.Call("Client.HandleCommunication", client.Id, &reply)
	if (err != nil){
		fmt.Println("There was an error trying to connect to the coordinator")
		fmt.Println("Invoking election")
		// code to invoke election here
	}
}


func (client *Client) Sendcandidacy(){ // invokeelection()
	fmt.Println("Discovery phase beginning...")
	for id, ip := range client.Clientlist{
		reply := Message{Type: "announce"}
		if id > client.Id{
			fmt.Println("Sending candidacy to", id)
			clnt, err := rpc.Dial("tcp", ip)
			if (err != nil){
				fmt.Println("Communication to", id, "failed.")
				continue
			}
			err = clnt.Call("Client.HandleCommunication", client.Id, &reply)
			if (err != nil){
				fmt.Println("Communication to", id, "failed")
				continue
			}
			if (reply.Type == "ack"){
				fmt.Println("Received a reply from", id)
				Higherid = true
			}
		}
	}
	if (Higherid == false){
		// function to make yourself coordinator
		client.Announcevictory()
	}
	Election_invoked = true
}

func (client *Client) Announcevictory(){ // Make yourself coordinator
	reply := Message{Type: "victory"}
	for id, ip := range client.Clientlist{
		clnt, err := rpc.Dial("tcp", ip)
		if (err != nil){
			fmt.Println("Communication to", id, "failed")
			continue
		}
		err = clnt.Call("Client.HandleCommunication", client.Id, &reply)
		if (reply.Type == "ack" && id > client.Id){
			// if you are announcing vicotry and you receive an ack message from a higher id, then just stop announcing
			fmt.Println("Message sent to", id, "successfully")
			fmt.Println("Client", id, "is awake")
			break
		}else{
			fmt.Println("For some reason, the", id, "did not set the new coordinator")
		}
	}
	return
}