package main

import (
    "fmt"
    "./networkModule/bcast"
    "./queueModule"
    "./driverModule/elevio"
    //"./networkModule/localip"
    "time"
	/*
    "reflect"
    "os"
    "net"
    "./queueModule"
    */
)


func main(){
	/*
	var local string
	local,_ = localip.LocalIP()
	fmt.Printf("%s\n", local)
	*/
	//queue.InitQueue()
	var test_order queue.OrderStruct
	test_order.Button = elevio.BT_HallUp
	test_order.Floor = 2
	test_order.ElevID = 1
	test_order.Cost = 5
	test_order.Cmd = queue.CostReq
	test_order.Timestamp = time.Now()
	//printOrder(test_order)

	trans_chan := make (chan queue.OrderStruct)
	rec_chan := make (chan queue.OrderStruct)
	go bcast.Transmitter(20005, trans_chan)
	go bcast.Receiver(20005, rec_chan)
	
	//go setstring(trans_chan)
	go setOrder(trans_chan)

	for{
		select{
		case rec := <- rec_chan:
			printOrder(rec)
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

func setOrder(trans chan<- queue.OrderStruct){
	var test_order queue.OrderStruct
	test_order.Button = elevio.BT_HallUp
	test_order.Floor = 0
	test_order.ElevID = 1
	test_order.Cost = 5
	test_order.Cmd = queue.CostReq
	test_order.Timestamp = time.Now()
	for{
		test_order.Floor+=1
		test_order.Timestamp = time.Now()
		trans <- test_order
		time.Sleep(5000* time.Millisecond)
	}

}

func printOrder(order queue.OrderStruct){
	fmt.Printf("Button: %d\nFloor: %d\nID: %d\nCost: %d\nCmd: %d\nTime: %s\n", order.Button, order.Floor, order.ElevID, order.Cost, order.Cmd, order.Timestamp.String())
}