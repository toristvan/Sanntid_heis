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

func IOwrapper(internal_new_order_chan chan<- config.OrderStruct, internal_floor_chan chan<- int){
  var new_order config.OrderStruct

  drv_floors  := make(chan int)
	drv_buttons := make(chan config.ButtonEvent)

  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollButtons(drv_buttons)

  for{
			select{
	      case button_input := <-drv_buttons:
	        new_order.ElevID 		= config.LocalID
	  			new_order.Button    = button_input.Button
	  			new_order.Floor     = button_input.Floor
	  			new_order.Cmd				= config.CostReq

	        internal_new_order_chan <- new_order

	      case floor_input := <- drv_floors: //kanskje unøvendig, ikke helt sikker
	        internal_floor_chan <- floor_input
	    }
  }
}

func main() {

	/*
	queue.InitQueue()
	elevio.Init("localhost:15657") //, num_floors)
	go elevclient.RunElevator()
	*/

	//var current_order config.OrderStruct //floorStatus
	var current_floor int
	//var elevEvent config.ButtonEvent

	//next_floor := elevEvent.Floor
	//order_type := elevEvent.Button

	//new_command         := make(chan config.ElevCommand)
	//status_elev_state   := make(chan config.Status)
	//sync_elev_state     := make(chan config.Status)

	internal_floor_chan := make(chan int)
	internal_new_order_chan 	:= make(chan config.OrderStruct)
	execute_chan	  				:= make(chan config.OrderStruct) //Receives first element in queue
	input_queue		  				:= make(chan config.OrderStruct)
	start_order_chan 				:= make(chan config.OrderStruct)
	add_order_chan 					:= make(chan config.OrderStruct)

	current_floor = fsm.ElevatorInit()
	Println(current_floor)
	queue.InitQueue()

	go IO.IOwrapper(internal_new_order_chan, internal_floor_chan)
	go queue.Queue(input_queue, execute_chan)
	go queue.DistributeOrder(start_order_chan, add_order_chan, config.LocalID)
	//start_order_chan <-chan config.OrderStruct
	//add_order_chan chan<- config.OrderStruct

	// go fsm.ElevStateMachine(executeOrder)

	for {
		//go fsm.ElevStateMachine(status_elev_state, sync_elev_state, order_type, next_floor)
		//go fsm.ElevInputCommand(new_command)

	  select {
		case floor_input := <- internal_floor_chan:
			Println(floor_input)

	  case new_order := <- internal_new_order_chan:
			//sende ordre til andre her
			input_queue <- new_order

/*
		case input := <-add_order_chan:
			Println("received new order")
			input_queue <- input

	  	var new_order config.OrderStruct
	    //sende ordre til andre her
	    new_order.Button     = button_input.Button
	    new_order.Floor      = button_input.Floor
	    new_order.Cost       = dummyCostFunc(button_input.Button, button_input.Floor)
	    new_order.ElevID     = localID
	    new_order.Timestamp  = time.Now()
	    input_queue <- new_order

	    case execute_order := <- pending_orders:
	    next_floor    = execute_order.Floor
	    order_type    = execute_order.Button
	    current_order = execute_order

		  //Add to watchdog here
		  elevio.SetButtonLamp(order_type, next_floor, true)

		  switch order_type {
		  case config.BT_HallUp, config.BT_HallDown:
		  	Println("Hall call")
		    Println("Floor:", current_order.Floor)

		    if next_floor < current_floor{
		    	new_command <- config.GoDown
		    } else if next_floor > current_floor {
		      new_command <- config.GoUp
		    } else {
		      new_command <- config.FloorReached
		    }
		 case config.BT_Cab:
		    Println("Cab call")
		    Println("Floor:", current_order.Floor)

		    if next_floor < current_floor{
		    	new_command <- config.GoDown
		    } else if next_floor > current_floor {
		      new_command <- config.GoUp
		    } else {
		      new_command <- config.FloorReached
		    }
		 }

		 sync_elev_state <- config.Active
		 */

	 /*
	 case floor_input := <- drv_floors:

	 	current_floor = floor_input
	  elevio.SetFloorIndicator(current_floor)

	  if current_floor == current_order.Floor {
	  	elevio.SetButtonLamp(current_order.Button, current_floor, false)
	    //queue.RemoveOrder(current_floor, localID)
	    new_command <- config.FloorReached
	  } else if current_floor < current_order.Floor && fsm.RetrieveElevState() == config.GoingUp {   //These two can be merged.
	    //Retransmit/reassign order
	  } else if current_floor > current_order.Floor && fsm.RetrieveElevState() == config.GoingDown {  //Readability tho?
	    //Retransmit/reassign order
	  }
	  sync_elev_state <- config.Active



	  case current_status := <- status_elev_state:
	  	switch current_status {
	    case config.Pending:
	       sync_elev_state <- config.Done
	    case config.Active:
	       sync_elev_state <- config.Pending
	    case config.Done:
	       sync_elev_state <- config.Active
	    }
		*/
	  }

	}

}
