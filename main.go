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

<<<<<<< HEAD
func IOwrapper(internal_new_order_chan chan<- config.OrderStruct, internal_floor_chan chan<- int){
	var new_order config.OrderStruct

	drv_floors  := make(chan int)
	drv_buttons := make(chan config.ButtonEvent)

	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollButtons(drv_buttons)

	for{
		select{
		case button_input := <-drv_buttons:
		    new_order.ElevID 	= config.LocalID
			new_order.Button    = button_input.Button
			new_order.Floor     = button_input.Floor
			new_order.Cmd		= config.CostReq

		    internal_new_order_chan <- new_order

	    case floor_input := <- drv_floors: //kanskje unøvendig, ikke helt sikker
	    	internal_floor_chan <- floor_input
	    }
	}
}

=======
>>>>>>> 196aef133bc1780e3ad8deef6df51c34b616a9d0
func main() {

	/*
	queue.InitQueue()
	elevio.Init("localhost:15657") //, num_floors)
	go elevclient.RunElevator()
	*/

	var current_floor int
	internal_floor_chan 			:= make(chan int)
	internal_new_order_chan 	:= make(chan config.OrderStruct)

	executed_order 						:= make(chan config.OrderStruct)
	execute_order							:= make(chan config.OrderStruct)
	execute_chan	  					:= make(chan config.OrderStruct) //Receives first element in queue
	executed_chan	  					:= make(chan config.OrderStruct)
	input_queue		  					:= make(chan config.OrderStruct)
	start_order_chan 					:= make(chan config.OrderStruct)
	add_order_chan 						:= make(chan config.OrderStruct)

	current_floor = fsm.ElevatorInit()
	Println(current_floor)
	queue.InitQueue()

	go IO.IOwrapper(internal_new_order_chan, internal_floor_chan)
	go fsm.ElevStateMachine(execute_order, executed_order, internal_floor_chan)
	go queue.Queue(input_queue, execute_chan, executed_chan)
	go queue.DistributeOrder(start_order_chan, add_order_chan, config.LocalID)

	for {
	  select {
	  case new_order := <- internal_new_order_chan:
			//sende ordre til andre her
			input_queue <- new_order

		case exe_ord := <- execute_chan:
			execute_order <- exe_ord

		case order_finished := <- executed_order:
			Println("finished")
			executed_chan <- order_finished

		/*
		case tmp := <-execute_chan:
			 execute_order <- tmp
		*/

	  }
	}

}
