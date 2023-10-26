# bully_algorithm

This question has been completely implemented in the `~/bully_algorithm` folder. So all commands listed below must be issued from there. Open up as many terminals as you want --> Each terminal will become a client when you run `go run main.go` from the parent directory. Go routines are used throughout the code to get a consensus across all the nodes, but the nodes themselves are not go routines, they are complete processes running in more than one terminal. This behaviour is in line with the requirements as confirmed with Professor Sudipta.

Another thing to note here is that, in some of the questions in this part, we are introducing some artificial delay using time.Sleep() to achieve some of the scenarios listed in the questions. We know that TCP orders messages in a queue. So this time.Sleep() tends to mess up the ordering of messages. So if you see the messages printed out of order, please do ignore it. Focus on the logic presented in the explanation. Rest assured, taking away the time.sleep() will print the messages in order as expected.

New nodes joining the network usually obtain a copy of the clientlist by contacting some arbitrary node. To simulate this behaviour, a `clientlist.json` has been included in the code repository. The clientlist is a data source for the newly spun up client to grab hold of an address from the address space. If the `clientlist.json` is is not empty when the first ever node in the network is spun up, you will see a message like this:

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/3d501093-61ac-427d-b354-aa0f232f719a)

However, shortly after it's spun up, the the "dead entry" of node 2 will be removed from the `clientlist.json`, so future nodes joining the network won't have to see the dead entry in the file. As such, the clientlist file is self-correcting. But, if you wish to start with a clean slate of nodes, then feel free to make sure that the `clientlist.json` file has an empty json string: 
```json
{}
```

Upon spinning up any node, you will also be presented with a menu as follows

```shell
Press 1 to enter something into my replica
Press 2 to view my replica
Press 3 to view my clientlist
```

If you are feeling old, please feel free to increase the `time.Sleep()` aka the timeout duration to something longer, so you can see the output clearly 😄. As of submission, it has been set to 1 seconds- so that the messages are sent and received on time. A longer timeout could result in TCP messages not reaching on time, or reaching after some other message. But, you may enter the sleep duration in the following methods to change the timout:

- InvokeReplicaSync() - coordinators detect failed nodes using this
- CoordinatorPing() - non-coordinators detect coordinator failures using this

The time.Sleep() duration in these two methods determines the timeout interval for failure detection, because that's how long the nodes wait for pinging each other. 

## Answers

### 1
In order to see the joint synchronization, open up 2-3 terminal for simplicity sake, and run `go run main.go`. You will see that each client now has a unique id assigned to it, and an IP address. You will also see a menu on each terminal. Press 2 to verify that the replica is empty in each terminal. Once done, on one of the terminals, press 1, and then enter the number you want to enter into the replica of that node. 
Client 0, 1, 2 started in three terminals respectively:
![image](https://github.com/fauzxan/distributed_systems/assets/92146562/86db6bba-f419-47af-8aac-b44a347f949f)

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/b843bfa6-8e2f-44c8-9feb-de86b0c35b2a)

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/14cea6b4-1b09-4c15-98bf-1a09c8ce1f67)

The server will synchronize once every 10 seconds. So you may enter any number of entries into the replicas in any of the terminals, and they will all be synchronized in one or two synchronziation cycles. In the SS below, there were some entries in node 0 and node 1 in the replica. They were synchronized within 2 cycles:

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/0b6f3115-a1e6-460e-924e-66dc68081140)

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/9073d449-4194-4434-8daf-b0f06ce3aaf6)

### 2
In order to measure the best and the worst case times, we make use of a timer to count the time elapsed between when the discover phase started in the clients that detect the election, and when the election ended. 
Four clients were spun up, and a failure was triggerred with the fourth client. The number of clients that detect the failure can be controlled by the controlling the timeout. It is possible that only one node detects the failure if the timeout is large enough. As mentioned, the timeout to detect the coordinator failure needs to be set in `CoordinatorPing()`. However, for the screenshots below, the timeout was only 1s. So multiple nodes detected the failure, and the time between discovery and victory receival (for non victors)/ victory sending(for victor) was calculated for all the nodes.

#### Worst case scenario:
This happens when the client with the lowest id is the one to detect the node failure. As shown below, the lowest id client took almost 3 ms to finish the election. 
![image](https://github.com/fauzxan/distributed_systems/assets/92146562/cbe40bb1-dbd8-4604-9234-332d15d5c1d8)

#### Best case scenario
This happens when the second highest id node is the one to detect the failure. This took around half the time as the highest id node:

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/ca4666ad-4a10-4ade-a6b9-231aadf58ef7)

The node with the middle id took more time than highest id, but less time than lowest id node:

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/6d804cf9-01ce-4437-9e99-2fff39e8e47d)



### 3
We will now consider a case where a node fails during the election. In order to achieve this, we will introduce a small delay when sending out victory messages, so we get sufficient time to fail 
a) the coordinator
b) non-coordinator node

The new coordinator stops two seconds before sending to the other node. This will give us enough time to fail the coordinator/ non-coordinator node. The delay has been introduced as follows:
```go
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
		time.Sleep(2 * time.Second) // <-- DELAY IS BEING INTRODUCED HERE
	}
	go client.InvokeReplicaSync()
}
```

#### Part a
This part will fail the coordinator after discovery, but before announcement. 

For this, we spun up four nodes in four different terminals. (You may do the same with 3, but I just want to demonstrate that this system works with more nodes as well)
In the screenshot given below, we can see that the last node, with id 3 just got created and is trying to announce to everyone that it is the coordinator. We can also see that the discovery phase is done. It is in the process of announcing to everyone. 

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/189eaa16-30c7-47c1-8ff1-114332f5a417)


Above, we see that the new coordinator has already sent messages to node 1 and node 2 who have acknowledge it as the coordinator. However, node 0 has not even received a victory message from node 3. So according to node 0, node 2 is the coordinator. But node 1 and node 2 think 3 is the coordinator. 

##### Node 2:
It was sending out periodic syncs-> then it received victory message from 3-> stopped sending out syncs

Because of the pinging algorithm, where nodes ping the coordinator periodically, this node was able to detect that the coordinator failed, and then started an election. 
![image](https://github.com/fauzxan/distributed_systems/assets/92146562/f565bb67-8958-4837-bc8d-2e1369188cdc)
##### Node 1: 

It was received periodic syncs-> then it received victory message from 3-> stopped sending out ping to 2 -> started sending out ping to 3

Again, the pinging failed, so it started election, and lost to 2. 
![image](https://github.com/fauzxan/distributed_systems/assets/92146562/b6218c18-99ff-47fc-b19a-cadef100afc2)
##### Node 0:
This node never knew any new coordinator other than 2. Hence it was, is, and continues to ping 2 only. In the middle it receives a victory message from 2, but this doesn't matter to 0, as it already believes that 2 is it's coordinator. 

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/0c6e115f-8ad9-4d34-96ea-e3c77401d9e0)

### Part b
In this part we will explore what happens when an arbitrary node fails during the election. As we shall see, the system I have built is self correcting. This means that the failed node will eventually be noticed by the elected coordinator, and the coordinator will take it out of its Clientlist. The other nodes will definitely detect that there is a failed node in the network as soon as they try to communicate with it. This is because all the `rcp.Dial()` instances have a code block that removes the node from their own Clientlist if the node is unable to contact the supposedly failed node. 
```go
clnt, err := rpc.Dial("tcp", coordinator_ip)
if (err != nil){
	fmt.Println("There was an error trying to connect to the coordinator")
	fmt.Println(err)
	fmt.Println("Invoking election")
	delete(client.Clientlist, client.Coordinator_id) // Delete from own clientlist
	client.updateclientlist() 
	go client.Sendcandidacy()
	return
}
err = clnt.Call("Client.HandleCommunication", send, &reply)
if (err != nil){
	fmt.Println("There was an error trying to connect to the coordinator")
	fmt.Println(err)
	fmt.Println("Invoking election")
	delete(client.Clientlist, client.Coordinator_id) // Delete from own clientlist
	go client.Sendcandidacy()
	client.updateclientlist() 
	return
}
```

However, as per requirements of the second part of the assignment, the non-coordinator nodes do not really communicate with each other, due to which the failed node's entry will continue to persist in the Clientlist of non-coordinator nodes. The non-coordinator nodes will remove them from their Clientlist as soon as they get into a situation where they start attempting to communicate with the failed node- i.e., when the non-coordinator node becomes the coordinator, it will try to `InvokeReplicaSync()` with everyone in it's Clientlist.  

In the image below, we see that the client 1 has failed after election started.

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/80797c63-7939-4dcd-bf1a-241dec48d1ee)

And, upon pressing `3` to view clientlist, we see that 1 is indeed not there in its clientlist anymore.

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/74519aa9-3c38-483a-b4df-cbd5c135cc42)


### 4
In order to see a scenario, where multiple clients start the election simultaneously, spin up three clients by running `go run main.go` in three separate clients simultaneously. Then kill the coordinator terminal by pressing `ctrl+c`. The discovery phase in the other two terminals would've begun almost at the same time (separated by a few milliseconds). Previously, in order to slow down the message output, the time.Sleep() value of the CoordinatorPing() method was set to 10 seconds. However, in order to bring about the behaviour required by this question, the time.Sleep() was set to 1 second. 
```go
func (client *Client) Coordinatorping(){ // communicatetocoordinator
	if (len(client.Clientlist) == 1){
		client.Coordinator_id = client.Id
		return
	}
	for {
		time.Sleep(1*time.Second) // <-- this line was modified
		// logic to dial coordinator tcp, and send message to the server.
	}
}
```

As such, we are able to observe from the following screenshots, that the two terminals simultaneously started election at `11:33:45:303`. 

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/445bb0ad-9f2a-4781-ae6a-2d2be2551081)

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/9a10ccfb-d768-461c-8c28-ea772c5a0dd3)


### 5
For an arbitrary node to leave the network, just hit `ctrl+c` in the terminal. 

![image](https://github.com/fauzxan/distributed_systems/assets/92146562/bbe0474f-1830-4c95-ac02-456fda28d8aa)

There are two cases:
1. If the arbitrary node is the coordinator, then the rest of them will figure out that the coordinator is down when trying to ping it. || All clients ping the coordinator every 10 seconds to see if it is alive.![image](https://github.com/fauzxan/distributed_systems/assets/92146562/b83b4b8b-3453-4359-a46e-2699c7fb8f25)

2. If the arbitrary node is not the coordinator, then the coordinator will detect the failure, and update it's clientlist. So next time it sends replica sync request, it won't send to the failed node. The rest of the nodes still have the failed node in their entry. But these nodes will remove the failed nodes as soon as they try and communicate with it. However, as per implementation, the non coordinator nodes never really communicate within themselves, therefore, they don't remove the failed node unless they themselves become the coordinator. ![image](https://github.com/fauzxan/distributed_systems/assets/92146562/6a61ecc7-0a42-49ce-97a8-33a086e8b1eb)
