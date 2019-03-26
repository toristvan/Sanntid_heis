package elevclient

import (
    "../driverModule/elevio"
    "../queueModule"
    "../fsmModule"
    "../configPackage"
    "fmt"
    "time"
)

type floorStatus struct{
  stop_up bool
  stop_down bool
}
var current_floor int
var stopArray[elevio.Num_floors] floorStatus

func initStopArray(){
  for i := elevio.Ground_floor ; i < elevio.Num_floors ; i++ {
    stopArray[i].stop_up = false
    stopArray[i].stop_down = false
  }
} 

func isEmpty(arr [elevio.Num_floors]floorStatus, from int, to int) bool{
  	for i := from ; i < to ; i++{
    	if arr[i].stop_up || arr[i].stop_down {
      		return false
    	}
  	}
  	return true
}

//An alert function which sends alert when elevator is idle when stopArray is not empty
func elevWakeUp(wakeup_chan chan<- bool){
  for {
    if !isEmpty(stopArray, elevio.Ground_floor, elevio.Num_floors) && (fsm.RetrieveElevState() == config.Idle) {
      //fmt.Println("Queue Alert!")
      wakeup_chan <- true
    }
    time.Sleep(1*time.Second)   //Unload CPU
  }
}

func executeOrder(execute_chan <-chan config.OrderStruct){ //, pending_orders chan<- floorStatus){
  for {

    new_order := <- execute_chan   //Input from queue

    switch new_order.Button{
    case config.BT_HallUp:
    	stopArray[new_order.Floor].stop_up = true
    case config.BT_HallDown:
    	stopArray[new_order.Floor].stop_down = true
    case config.BT_Cab:
    	stopArray[new_order.Floor].stop_up = true
    	stopArray[new_order.Floor].stop_down = true

    }
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



func ElevRunner(trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct){
  //var current_dir config.MotorDirection
  //var prev_dir config.MotorDirection
  //fsm.RetrieveState()
  var current_state config.ElevStateType = config.Idle

  source_trans_chan := make (chan config.OrderStruct,10)
  sink_rec_chan     := make (chan config.OrderStruct,10)

  floor_chan    := make (chan int)
  input_queue   := make (chan config.OrderStruct)
  execute_chan  := make (chan config.OrderStruct)
  wakeup_chan   := make (chan bool)
  elev_cmd_chan := make (chan config.ElevCommand)

  go executeOrder(execute_chan)
  go queue.Queue(input_queue, execute_chan, source_trans_chan, sink_rec_chan)

  go IOwrapper(input_queue, floor_chan)
  go elevWakeUp(wakeup_chan)
  go fsm.ElevStateMachine(elev_cmd_chan)

  for{
    select{
    case tmp := <- source_trans_chan:
      fmt.Println("tx elevclient",tmp)
      trans_main_chan <- tmp

    case tmp := <- rec_main_chan:
      fmt.Println("rx elevclient",tmp)
      sink_rec_chan <- tmp

    
    case current_floor = <- floor_chan:
    	current_state = fsm.RetrieveElevState()
      elevio.SetFloorIndicator(current_floor)
      
    	switch current_state{
        case config.GoingUp:
        	if (stopArray[current_floor].stop_up) || (stopArray[current_floor].stop_down && isEmpty(stopArray, current_floor+1, elevio.Num_floors)) {
	        	//Stop routine
	        	elev_cmd_chan <- config.FloorReached
	        	stopArray[current_floor].stop_up = false
	        	stopArray[current_floor].stop_down = false
	        	queue.RemoveOrder(current_floor, config.LocalID)
        		fmt.Println(stopArray)
        		//for i := 0; i < 3; i++ {
        			//elevio.SetButtonLamp(config.ButtonType(i), current_floor, false)  //Switch off all lights associated with floor
        		//}

        	}  
        	//Stop again if new order received when at floor with open door 
        	if fsm.RetrieveElevState() == config.AtFloor && (stopArray[current_floor].stop_up || stopArray[current_floor].stop_down) {
        		elev_cmd_chan <- config.FloorReached
	        	stopArray[current_floor].stop_up = false
	        	stopArray[current_floor].stop_down = false
	        	queue.RemoveOrder(current_floor, config.LocalID)
        		//fmt.Println(stopArray)
        	}

        	if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
            	//fmt.Println(current_state) //Remove prints
            	elev_cmd_chan <- config.GoUp
          	} else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
            	elev_cmd_chan <- config.GoDown
          	} else{
            	elev_cmd_chan <- config.Finished
          	}
        case config.GoingDown:
        	if (stopArray[current_floor].stop_down) || (stopArray[current_floor].stop_up && isEmpty(stopArray, elevio.Ground_floor, current_floor)) {
				//Stop routine
        		elev_cmd_chan <- config.FloorReached
        		stopArray[current_floor].stop_up = false
        		stopArray[current_floor].stop_down = false
	        	queue.RemoveOrder(current_floor, config.LocalID)
        		//fmt.Println(stopArray)        		

	        	
        		//for i := 0; i < 3; i++ {
        			//elevio.SetButtonLamp(config.ButtonType(i), current_floor, false)  //Switch off all lights associated with floor
        		//}
        	}
        	//Stop again if new order received when at floor with open door 
        	if fsm.RetrieveElevState() == config.AtFloor && (stopArray[current_floor].stop_up || stopArray[current_floor].stop_down) {
        		elev_cmd_chan <- config.FloorReached
	        	stopArray[current_floor].stop_up = false
	        	stopArray[current_floor].stop_down = false
	        	queue.RemoveOrder(current_floor, config.LocalID)
        		//fmt.Println(stopArray)
        	}
        	if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
            	elev_cmd_chan <- config.GoDown
          	} else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
            	elev_cmd_chan <- config.GoUp
          	} else{
            	elev_cmd_chan <- config.Finished
          	}
        }
        

    case <- wakeup_chan:       //Channel dedicated to alert if elevator is idle with orders in stopArray
     	if stopArray[current_floor].stop_up || stopArray[current_floor].stop_down {
	        elev_cmd_chan <- config.FloorReached
	        stopArray[current_floor].stop_up = false
	        stopArray[current_floor].stop_down = false
	        queue.RemoveOrder(current_floor, config.LocalID)

	        elev_cmd_chan <- config.Finished
	        //Turn off lights should be in RemoveArray instead
	        
	        //for i := 0; i < 3; i++ {
	          //elevio.SetButtonLamp(config.ButtonType(i), current_floor, false)  //Switch off all lights associated with floor
	        //}
	    } else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
        	elev_cmd_chan <- config.GoDown
      	} else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
        	elev_cmd_chan <- config.GoUp
      	} else {
        	elev_cmd_chan <- config.Finished
      	}
    }
  }
}