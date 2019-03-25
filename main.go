package main

import (
	"./driverModule/elevio"
	"./fsmModule"
	"./queueModule"
	"./configPackage"
	"./IOModule"
	."fmt"
)

type floorStatus struct{
	 stop bool
	 Button config.ButtonType
}

var stopArray[elevio.Num_floors] floorStatus
	//var index = 1
func  isEmpty(arr [elevio.Num_floors]floorStatus, from int, to int) bool{
	  for i := from - 1 ; i < to - 1 ; i++{
	    if arr[i].stop {
	      return false
	    }
	  }
	  return true
}
func dummyCostFunc(hallCall config.OrderStruct) int {
	 return 1
}
func executeOrder(execute_chan <-chan config.OrderStruct, pending_orders chan<- floorStatus){
	  select{
	  case new_order := <- execute_chan:   //Input from queue
	    stopArray[new_order.Floor-1].Button = new_order.Button
	    stopArray[new_order.Floor-1].stop = true

	    pending_orders <- stopArray[new_order.Floor]
	 	}
}

func main() {

	/*
	queue.InitQueue()
	elevio.Init("localhost:15657") //, num_floors)
	go elevclient.RunElevator()
	*/

	var current_floor int
	internal_floor_chan 			:= make(chan int)
	internal_new_order_chan 	:= make(chan config.OrderStruct)

	finished := make(chan bool)
	execute_order							:= make(chan config.OrderStruct)
	execute_chan	  					:= make(chan config.OrderStruct) //Receives first element in queue
	input_queue		  					:= make(chan config.OrderStruct)
	start_order_chan 					:= make(chan config.OrderStruct)
	add_order_chan 						:= make(chan config.OrderStruct)

	current_floor = fsm.ElevatorInit()
	Println(current_floor)
	queue.InitQueue()

	go IO.IOwrapper(internal_new_order_chan, internal_floor_chan)
	go fsm.ElevStateMachine(execute_order, internal_floor_chan, finished)
	go queue.Queue(input_queue, execute_chan)
	go queue.DistributeOrder(start_order_chan, add_order_chan, config.LocalID)

	for {
	  select {
	  case new_order := <- internal_new_order_chan:
			//sende ordre til andre her
			input_queue <- new_order

		case exe_ord := <- execute_chan:
			execute_order <- exe_ord

		case <- finished:
			Println("finished")

		/*
		case tmp := <-execute_chan:
			 execute_order <- tmp
		*/

	  }
	}

}
