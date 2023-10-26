package client

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"os"
	"sync"
	"time"
)


type Client struct{
	Lock sync.Mutex
	Id int
	Coordinator_id int
	Clientlist map[int]string
	Replica []int
}

// Flag variables:
var Higherid = false

// Message types
var ANNOUNCE = "announce"
var SYNC = "sync"
var CORRECTION = "correction"
var ACK = "ack"
var VICTORY = "victory"
var PING = "ping"

// Timer variable for timing elections:
var Timer = time.Now()

func (client *Client) HandleCommunication(message *Message, reply *Message) error{ // Handle communication
	client.Lock.Lock()
	defer client.Lock.Unlock()

	if (message.Type == SYNC){
		fmt.Println("Received a sync request from", message.From)
		reply.Replica = client.Replica
	}else if(message.Type == ANNOUNCE){
		fmt.Println("Received an announcement from", message.From)
		if (client.Check(message.From)){
			fmt.Println("Adding client into clientlist")
			client.Clientlist[message.From] = message.IpMapping
		}
		reply.Type = ACK
		go client.Sendcandidacy()
	}else if(message.Type == VICTORY){
		fmt.Println("Received a victory message from", message.From)
		client.Coordinator_id = message.From
		client.Clientlist = message.Clientlist
		fmt.Println(client.Coordinator_id, "is now my coordinator")
		reply.Type = ACK
		go client.Coordinatorping()
		// Election ended here, so I must print out time elapsed till now
		fmt.Println("********************************")
		fmt.Println("Time elapsed since election start:", time.Since(Timer))
		fmt.Println("********************************")
	}else if(message.Type == CORRECTION){
		fmt.Println("\nReceived a correction message from", message.From, client.Replica)
		client.Replica = message.Replica
		reply.Type = ACK
	}else if(message.Type == PING){
		// fmt.Println("Received a ping from", message.From)
		reply.Type = ACK
	}
	return nil
}

func (client *Client) Coordinatorping(){ // communicatetocoordinator
	if (len(client.Clientlist) == 1){
		client.Coordinator_id = client.Id
		return
	}
	for {
		time.Sleep(1*time.Second)
		if (client.Id == client.Coordinator_id){return} // Don't ping if I am the coordinator
		coordinator_ip := client.Clientlist[client.Coordinator_id]
		reply := Message{}
		send := Message{Type: PING, From: client.Id}
		clnt, err := rpc.Dial("tcp", coordinator_ip)
		if (err != nil){
			fmt.Println("There was an error trying to connect to the coordinator")
			fmt.Println(err)
			fmt.Println("Invoking election")
			delete(client.Clientlist, client.Coordinator_id)
			client.updateclientlist()
			go client.Sendcandidacy()
			return
		}
		err = clnt.Call("Client.HandleCommunication", send, &reply)
		if (err != nil){
			fmt.Println("There was an error trying to connect to the coordinator")
			fmt.Println(err)
			fmt.Println("Invoking election")
			delete(client.Clientlist, client.Coordinator_id)
			go client.Sendcandidacy()
			client.updateclientlist()
			return
		}
		if (reply.Type == ACK){
			// fmt.Println("Pinged the coordinator", client.Coordinator_id,"It's alive")
			fmt.Println("")
		}
	}
}

func (client *Client) InvokeReplicaSync(){
	for {
		time.Sleep(1*time.Second)
		if client.Id != client.Coordinator_id {return} // Don't start syncing if I am not the coordinator
		fmt.Println("\nStarting replica sync")
		for id := range client.Clientlist{
			if (id == client.Id){continue}
			send := Message{
				Type: SYNC,
				From: client.Id,
			}
			reply := Message{}
			fmt.Println("Sending sync message to", id)
			ip := client.Clientlist[id]
			clnt, err := rpc.Dial("tcp", ip)
			if(err != nil){
				fmt.Println("Client", id, "seems to be down. Will try again in next sync", err)
				delete(client.Clientlist, id)
				client.updateclientlist()
				continue
			}
			err = clnt.Call("Client.HandleCommunication", send, &reply)
			if(err != nil){
				fmt.Print("Error receiving a response", err)
				delete(client.Clientlist, id)
				client.updateclientlist()
				continue
			}
			client.Lock.Lock()
			client.Replica = append(client.Replica, reply.Replica... )
			client.Lock.Unlock()
		}
		client.Replica = client.removeDuplicates(client.Replica)
		fmt.Println("Updated client replica:", client.Replica)
		for id := range client.Clientlist{
			if (id == client.Id){continue}
			send := Message{
				Type: CORRECTION,
				From: client.Id,
				Replica: client.Replica,
			}
			reply := Message{}
			fmt.Println("Sending correction message to", id)
			ip := client.Clientlist[id]
			clnt, err := rpc.Dial("tcp", ip)
			if(err != nil){
				fmt.Println("Client", id, "seems to be down.	", err)
				continue
			}
			err = clnt.Call("Client.HandleCommunication", send, &reply)
			if(err != nil){
				fmt.Print("Error receiving a response", err)
				continue
			}
			if reply.Type == ACK{
				fmt.Println("Client", id, "successfully synchronized!")
			}
		}
	}
}


func (client *Client) Sendcandidacy(){ // invokeelection()
	Timer = time.Now() //  This is when I started the election, elapsed time is marked throughout the code
	fmt.Println("Discovery phase beginning at", Timer.Local().UTC())
	var Higherid = false
	for id, ip := range client.Clientlist{
		reply := Message{}
		send := Message{Type: ANNOUNCE, From: client.Id, IpMapping: client.Clientlist[client.Id]}
		if id > client.Id{
			fmt.Println("Sending candidacy to", id)
			clnt, err := rpc.Dial("tcp", ip)
			if (err != nil){
				fmt.Println("Communication to", id, "failed.")
				delete(client.Clientlist, id)
				go client.updateclientlist()
				continue
			}
			err = clnt.Call("Client.HandleCommunication", send, &reply)
			if (err != nil){
				fmt.Println("Communication to", id, "failed")
				delete(client.Clientlist, id)
				go client.updateclientlist()
				continue
			}
			if (reply.Type == "ack"){
				fmt.Println("Received an ACK reply from", id)
				Higherid = true
			}
		}
	}
	if (!Higherid){
		// function to make yourself coordinator
		client.Coordinator_id = client.Id
		go client.Announcevictory()
	}else{
		go client.Coordinatorping()
	}
}

func (client *Client) Announcevictory(){ // Make yourself coordinator
	send := Message{Type: VICTORY, From: client.Id, Clientlist: client.Clientlist}
	reply := Message{}
	fmt.Println("No higher id node found. I am announcing victory! Current clientlist (you may fail clients now):", client.Clientlist)
	client.Printclients()
	for id, ip := range client.Clientlist{
		if (id == client.Id){continue}
		clnt, err := rpc.Dial("tcp", ip)
		if (err != nil){
			fmt.Println("Communication to", id, "failed")
			delete(client.Clientlist, id)
			go client.updateclientlist()
			continue
		}
		err = clnt.Call("Client.HandleCommunication", send, &reply)
		if (err != nil){
			fmt.Println(err)
			continue
		}
		fmt.Println("Victory message sent to", id, "at", ip)
		if (reply.Type == "ack" && id > client.Id){
			// if you are announcing vicotry and you receive an ack message from a higher id, then just stop announcing
			fmt.Println("Message sent to", id, "successfully")
			fmt.Println("Client", id, "is awake")
			break
		}else if(reply.Type == ACK){
			fmt.Println("Client", id, "acknowledged me as coordinator")
		}
		// time.Sleep(2 * time.Second) // uncomment this line to introduce delay, so you can fail some nodes in the meantime
	}
	// This is where we calculate the elapsed time in case the client wins the election.
	fmt.Println("********************************")
	fmt.Println("Time elapsed since election start:", time.Since(Timer))
	fmt.Println("********************************")
	go client.InvokeReplicaSync()
}


func (client *Client) EnterReplica(num int){
	client.Replica = append(client.Replica, num)
	fmt.Println("Entered number into local replica. Replica is now", client.Replica)
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
	for id := range client.Clientlist {		
		if (id > max){max = id}
	}
	return max
}

func (client *Client) Check(n int) bool{
	/*
		true means that id is not taken | id is not there in the clientlist
		false means that id is taken | id is there in the clientlist
	*/
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
		if (check){return i}
	}
}

func (client *Client) removeDuplicates(array []int) []int {
	client.Lock.Lock()
	defer client.Lock.Unlock()
    seen := map[int]bool{}

    for _, element := range array {
        if _, ok := seen[element]; !ok {
            seen[element] = true
        }
    }
    uniqueArray := []int{}
    for element := range seen {
        uniqueArray = append(uniqueArray, element)
    }
    return uniqueArray
}

func (client *Client) updateclientlist(){
	jsonData, err := json.Marshal(client.Clientlist)
	if err != nil {
		fmt.Println("Could not marshal into json data")
		panic(err)
	}
	err = os.WriteFile("clientlist.json", jsonData, os.ModePerm)
	if err != nil {
		fmt.Println("Could not write into file")
		panic(err)
	}
}