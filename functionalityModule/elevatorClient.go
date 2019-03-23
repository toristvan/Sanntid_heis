package elevclient

import (
    "./../driverModule/elevio"
    "./../fsmModule"
    "./../queueModule"
    "./../configPackage"
    ."fmt"
	  "time"
)

const localID = 1

func dummyCostFunc(hallCall config.ButtonType, floor int) int {
  return 1
}

func RunElevator(){

    var current_order config.OrderStruct
    var current_floor int

    var elevEvent config.ButtonEvent
    next_floor := elevEvent.Floor
    order_type := elevEvent.Button

    new_command         := make(chan config.ElevCommand)
    status_elev_state   := make(chan config.Status)
    sync_elev_state     := make(chan config.Status)

    drv_buttons := make(chan config.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)

    order_chan     := make(chan config.OrderStruct)
    pending_orders := make(chan config.OrderStruct, 5)

    current_floor = fsm.ElevatorInit()
    queue.InitQueue()

    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go fsm.ElevStateMachine(status_elev_state, sync_elev_state, order_type, next_floor)
    go fsm.ElevInputCommand(new_command)
    go queue.Queue(order_chan)    //Second channel input_queue?

    for {

	    select {
	    case button_input := <-drv_buttons:
	        var new_order config.OrderStruct

	          //sende ordre til andre her
	    	new_order.Button     = button_input.Button
	    	new_order.Floor      = button_input.Floor
	        new_order.ElevID     = dummyCostFunc(button_input.Button, button_input.Floor)
	        new_order.Timestamp  = time.Now()
	    	//input_queue <- new_order   no use for separate channel?
	    	order_chan <- new_order

	    case new_order := <- order_chan:   //Input from queue
	        Println("new_order")
	        if new_order.ElevID == localID{
	          pending_orders <- new_order
	        }

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

	    case floor_input := <- drv_floors:

	        current_floor = floor_input
	        elevio.SetFloorIndicator(current_floor)
	        Println("Current floor:", current_floor)

	        if current_floor == current_order.Floor {
	            elevio.SetButtonLamp(current_order.Button, current_floor, false)
	            //queue.RemoveOrder(current_floor, localID)
	            new_command <- config.FloorReached
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
      	}
    }
}
