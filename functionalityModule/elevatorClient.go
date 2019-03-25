package elevclient

import (
    "../driverModule/elevio"
    "../queueModule"
    "../fsmModule"
    "../configPackage"
    "fmt"
    "time"
)
/*
=========================== Bothause! ================================
Denne fungerer med å bruke "Test queue" i queueModule (kanskje, har blitt modifisert)

// TODO:
Legge inn retransmit order hvis en blir passert. Line 142->
*/

type floorStatus struct{
  stop bool
  Button config.ButtonType
}

var current_floor int
//convert to int
var stopArray[elevio.Num_floors] floorStatus

func initStopArray(){
  for i := elevio.Ground_floor ; i < elevio.Num_floors ; i++ {
    stopArray[i].stop = false
    stopArray[i].Button = config.BT_Cab
  }
} 

//var index = 1

func  isEmpty(arr [elevio.Num_floors] floorStatus, from int, to int) bool{
  
  if from < elevio.Ground_floor{
  	fmt.Printf("Index out of bounds\n")
  	from = elevio.Ground_floor
  }
  if to > elevio.Num_floors{
  	fmt.Printf("Index out of bounds\n")
  	to = elevio.Num_floors
  }
  for i := from ; i < to ; i++{
    if arr[i].stop {
      return false
    }
  }
  return true
}

//An alert function which sends alert when elevator is idle when stopArray is not empty
func queueAlert(alert_chan chan<- bool){
  for {
    if !isEmpty(stopArray, elevio.Ground_floor, elevio.Num_floors) && (fsm.RetrieveElevState() == config.Idle/* || fsm.RetrieveElevState() == config.AtFloor*/){
      fmt.Println("Queue Alert!")
      alert_chan <- true
    }
    time.Sleep(1*time.Second)   //Unload CPU
  }
}

func dummyCostFunc(hallCall config.OrderStruct) int {
  return 1
}

func executeOrder(execute_chan <-chan config.OrderStruct){ //, pending_orders chan<- floorStatus){
  for {
    //select{
    //case 
    new_order := <- execute_chan   //Input from queue

    stopArray[new_order.Floor].Button = new_order.Button
    stopArray[new_order.Floor].stop = true
    fmt.Println(stopArray)
    queue.RemoveOrder(new_order.Floor, config.LocalID)   //Remove order when arriving instead of here

  //pending_orders <- stopArray[index]
  }
}

func IOwrapper(new_order_chan chan<- config.OrderStruct, floor_chan chan<- int){

  drv_floors  := make(chan int)
  drv_buttons := make(chan config.ButtonEvent)

  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollButtons(drv_buttons)

  elevio.SetMotorDirection(config.MD_Down)
  current_floor = <- drv_floors
  elevio.SetMotorDirection(config.MD_Stop)
  elevio.SetFloorIndicator(current_floor)

  for{
      select{
      case button_input := <-drv_buttons:
        var new_order config.OrderStruct
        new_order.ElevID    = config.LocalID
        new_order.Button    = button_input.Button
        new_order.Floor     = button_input.Floor

        if (new_order.Button != config.BT_Cab){
          new_order.Cmd = config.CostReq
        } else{
          new_order.Cmd = config.OrdrAdd
        }

        new_order_chan <- new_order

    case current_floor = <- drv_floors: //kanskje unøvendig, ikke helt sikker
        floor_chan <- current_floor
    }
  }
}

func ElevRunner(){
	//var current_dir config.MotorDirection
	//var prev_dir config.MotorDirection
	//var current_state config.ElevStateType = config.Idle
	//fsm.RetrieveState()
	var prev_state config.ElevStateType = config.Idle


	floor_chan  := make (chan int)
	input_queue := make (chan config.OrderStruct)
	execute_chan := make (chan config.OrderStruct)
	alert_chan := make(chan bool)
	elev_cmd_chan := make (chan config.ElevCommand)

	go executeOrder(execute_chan)
	go queue.Queue(input_queue, execute_chan)
	go IOwrapper(input_queue, floor_chan)
	go queueAlert(alert_chan)
	go fsm.ElevStateMachine(elev_cmd_chan)

	for{
	    select{
	    case new_floor := <- floor_chan:
	    	prev_state = fsm.RetrieveElevState()   
	    	//can just make current_floor = floor_chan?   
	    	current_floor = new_floor
	    	if stopArray[current_floor].stop{
	        	//Stop routine
	        	elev_cmd_chan <- config.FloorReached
	        	stopArray[new_floor].stop = false
	        	for i := 0; i < 3; i++ {
	        		elevio.SetButtonLamp(config.ButtonType(i), new_floor, false)  //Switch off all lights associated with floor
	        	}
	        	fmt.Println(stopArray)
	      	}
	      	elevio.SetFloorIndicator(current_floor)
	      	//elev_cmd_chan <- config.Finished
	      
	      switch prev_state{
	        case config.GoingUp:
	          if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
	            fmt.Println(prev_state)
	            elev_cmd_chan <- config.GoUp
	          } else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
	            elev_cmd_chan <- config.GoDown
	          } else{
	            elev_cmd_chan <- config.Finished
	          }
	        case config.GoingDown:
	          if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
	            elev_cmd_chan <- config.GoDown
	          } else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
	            elev_cmd_chan <- config.GoUp
	          } else{
	            elev_cmd_chan <- config.Finished
	          }
	        }
	        

	    case <- alert_chan:       //Channel dedicated to alert if elevator is idle with orders in stopArray
	    	
	    	if !isEmpty(stopArray, elevio.Ground_floor, current_floor){ //If orders below
	        	elev_cmd_chan <- config.GoDown
	    	} else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){ //Elseif order above
	        	elev_cmd_chan <- config.GoUp
	    	} else if stopArray[current_floor].stop { //Elseif order at current floor
	        	elev_cmd_chan <- config.FloorReached
	        	stopArray[current_floor].stop = false
	        	//elev_cmd_chan <- config.Finished //Needed in order to enter Idle - bad workaround?
	        	//Turn off lights should be in RemoveArray instead
	        	for i := 0; i < 3; i++ {
	          		elevio.SetButtonLamp(config.ButtonType(i), current_floor, false)  //Switch off all lights associated with floor
	        	}
	    	} else {
	        	elev_cmd_chan <- config.Finished
	    	}
	    	
	    }
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

