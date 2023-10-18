package client

import "fmt"

type Message struct{
	Type string // ack | sync | announce | victory
	Replica []int
}

func (message *Message) Printmsg(){
	fmt.Println("Message: ", message.Type)
}