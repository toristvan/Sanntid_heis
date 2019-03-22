package main

import (
    "fmt"
    "./networkModule/bcast"
    "./queueModule"
    //"./networkModule/localip"
    "time"
	/*
    "reflect"
    "os"
    "net"
    "./driverModule/elevio"
    "./queueModule"
    */
)


func main(){
	/*
	var local string
	local,_ = localip.LocalIP()
	fmt.Printf("%s\n", local)
	*/
	queue.InitQueue()
	trans_chan := make (chan string)
	rec_chan := make (chan string)
	go bcast.Transmitter(20005, trans_chan)
	go bcast.Receiver(20005, rec_chan)
	
	go setstring(trans_chan)
	for{
		select{
		case rec := <- rec_chan:
			fmt.Printf("Received: %s\n", rec)
			}

	}

	
}

func setstring(trans chan<- string){	
	var num int = 0
	for {
		var send string = fmt.Sprintf("sending: %d", num)
		trans <- send
		num+=1
		time.Sleep(1000 * time.Millisecond)
		//fmt.Printf(send)
	}
}