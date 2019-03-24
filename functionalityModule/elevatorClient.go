package elevclient

import (
    "./../driverModule/elevio"
    "./../queueModule"
    "./../fsmModule"
    "./../configPackage"
    //."fmt"
	  //"time"
)
/*
=========================== Bothause! ================================
Denne fungerer med å bruke "Test queue" i queueModule (kanskje, har blitt modifisert)

// TODO:
Legge inn retransmit order hvis en blir passert. Line 142->
*/

const localID = 1

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


func executeOrder(execute_chan <-chan config.OrderStruct){ //, pending_orders chan<- floorStatus){
  select{
  case new_order := <- execute_chan:   //Input from queue
    stopArray[new_order.Floor-1].Button = new_order.Button
    stopArray[new_order.Floor-1].stop = true

    //pending_orders <- stopArray[index]
  }
}

/*
func RunElevator(){

    var current_order config.OrderStruct //floorStatus
    var current_floor int

    var elevEvent config.ButtonEvent
    next_floor := elevEvent.Floor
    order_type := elevEvent.Button


    new_command         := make(chan config.ElevCommand)
    status_elev_state   := make(chan config.Status)
    sync_elev_state     := make(chan config.Status)

    drv_floors  := make(chan int)
    //drv_buttons := make(chan config.ButtonEvent)
    //drv_obstr   := make(chan bool)
    //drv_stop    := make(chan bool)

    execute_chan	  := make(chan config.OrderStruct) //Receives first element in queue
    input_queue		  := make(chan config.OrderStruct)
    pending_orders  := make(chan floorStatus, 5) //Why 5?

    current_floor = fsm.ElevatorInit()
    queue.InitQueue()

    go elevio.PollFloorSensor(drv_floors)
    //go elevio.PollButtons(drv_buttons)
    //go elevio.PollObstructionSwitch(drv_obstr)
	  //go elevio.PollStopButton(drv_stop)
	  
    go fsm.ElevStateMachine(status_elev_state, sync_elev_state, order_type, next_floor)
    go fsm.ElevInputCommand(new_command)
    go queue.Queue(input_queue, execute_chan)
    go executeOrder(execute_chan, pending_orders)

    for {

      	select {
        //Why is this here?
      	case button_input := <-drv_buttons:
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

      	}

    }
}
*/

func ElevRunner(){
  
  var current_floor int
  //var current_dir config.MotorDirection
  //var prev_dir config.MotorDirection
  var current_state config.ElevStateType
  var prev_state config.ElevStateType


  drv_floors  := make (chan int)
  input_queue := make (chan config.OrderStruct)
  execute_chan := make (chan config.OrderStruct)
  //elev_state_chan := make (chan config.ElevStateType)
  elev_cmd_chan := make (chan config.ElevCommand)


  go elevio.PollFloorSensor(drv_floors)
  go queue.Queue(input_queue, execute_chan)
  //go executeOrder(execute_chan)
  go fsm.ElevStateMachine2(elev_cmd_chan, &current_state)



  for{
    select{
    case new_floor := <- drv_floors:
      prev_state = current_state
      current_floor = new_floor
      if stopArray[new_floor].stop{
        //Stop routine
        elev_cmd_chan <- config.FloorReached 
        stopArray[new_floor].stop = false
      }
      elevio.SetFloorIndicator(current_floor)
      
      switch prev_state{
      case config.GoingUp:
        if !isEmpty(stopArray, current_floor, elevio.Num_floors){
          elev_cmd_chan <- config.GoUp
        } else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
          elev_cmd_chan <- config.GoDown
        } else{
          elev_cmd_chan <- config.Finished
        }
      case config.GoingDown:
        if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
          elev_cmd_chan <- config.GoDown
        } else if !isEmpty(stopArray, current_floor, elevio.Num_floors){
          elev_cmd_chan <- config.GoUp
        } else{
          elev_cmd_chan <- config.Finished
        }
      }
    }
  }
}
