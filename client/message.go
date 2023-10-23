package client

import "fmt"

type Message struct{
	Type string // ack | sync | announce | victory
	Replica []int
	From int
	Clientlist map[int]string
	IpMapping string
}

func (message *Message) Printmsg(){
	fmt.Println("Message: ", message.Type)
	if (message.Replica != nil){fmt.Println(message.Replica)}
}