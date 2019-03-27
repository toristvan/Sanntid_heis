package elevclient

import (
    "../driverModule/elevio"
    "../queueModule"
    "../fsmModule"
    "../configPackage"
    //"fmt"
    "time"
    //"math"
)

type floorStatus struct{
  stop_up bool
  stop_down bool
}

//var config.Current_floor int
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

func ExecuteOrder(execute_chan <-chan config.OrderStruct){ //, pending_orders chan<- floorStatus){
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

func IOwrapper(raw_order_chan chan<- config.OrderStruct){
  drv_buttons := make(chan config.ButtonEvent)
  go elevio.PollButtons(drv_buttons)

  for{
      button_input := <-drv_buttons
      var new_order config.OrderStruct
      new_order.ElevID    = config.LocalID
      new_order.Button    = button_input.Button
      new_order.Floor     = button_input.Floor
      new_order.MasterID  = config.LocalID

      if (new_order.Button != config.BT_Cab){
        new_order.Cmd = config.CostReq
      } else{
        new_order.Cmd = config.OrdrAdd
      }

      raw_order_chan <- new_order
  }
}

func setFloorFalse() config.OrderStruct{
  var order_to_delete config.OrderStruct

  order_to_delete.Floor = config.Current_floor
  order_to_delete.ElevID = config.LocalID
  order_to_delete.Cmd = config.OrdrDelete

  stopArray[config.Current_floor].stop_up = false
  stopArray[config.Current_floor].stop_down = false
  queue.RemoveOrder(config.Current_floor, config.LocalID)
  return order_to_delete

}


func ElevRunner(elev_cmd_chan chan<- config.ElevCommand, delete_order_chan chan<- config.OrderStruct /*, trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct*/){

  var current_state config.ElevStateType = config.Idle

  drv_floors  := make(chan int)
  wakeup_chan   := make (chan bool)

  go elevWakeUp(wakeup_chan)
  go elevio.PollFloorSensor(drv_floors)

  for{
    select{
    case config.Current_floor = <- drv_floors:
      current_state = fsm.RetrieveElevState()
      elevio.SetFloorIndicator(config.Current_floor)

      switch current_state{
        case config.GoingUp:
          if (stopArray[config.Current_floor].stop_up) || (stopArray[config.Current_floor].stop_down && isEmpty(stopArray, config.Current_floor+1, elevio.Num_floors)) {
            //Stop routine
            elev_cmd_chan <- config.FloorReached
            delete_order_chan <- setFloorFalse()
        	}
        	//Stop again if new order received when at floor with open door
        	if fsm.RetrieveElevState() == config.AtFloor && (stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down) {
        	 	elev_cmd_chan <- config.FloorReached
	        	delete_order_chan <- setFloorFalse()
        	}

        	if !isEmpty(stopArray, config.Current_floor+1, elevio.Num_floors){
            	elev_cmd_chan <- config.GoUp
          } else if !isEmpty(stopArray, elevio.Ground_floor, config.Current_floor){
            	elev_cmd_chan <- config.GoDown
          } else{
            	elev_cmd_chan <- config.Finished
          }

        case config.GoingDown:
        	if (stopArray[config.Current_floor].stop_down) || (stopArray[config.Current_floor].stop_up && isEmpty(stopArray, elevio.Ground_floor, config.Current_floor)) {
				    //Stop routine
        		elev_cmd_chan <- config.FloorReached
        		delete_order_chan <- setFloorFalse()
        	}
        	//Stop again if new order received when at floor with open door
        	if fsm.RetrieveElevState() == config.AtFloor && (stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down) {
        		elev_cmd_chan <- config.FloorReached
	        	delete_order_chan <- setFloorFalse()

        	}

        	if !isEmpty(stopArray, elevio.Ground_floor, config.Current_floor){
            	elev_cmd_chan <- config.GoDown
          } else if !isEmpty(stopArray, config.Current_floor+1, elevio.Num_floors){
            	elev_cmd_chan <- config.GoUp
          } else{
            	elev_cmd_chan <- config.Finished
          }
        }
    case <- wakeup_chan:       //Channel dedicated to alert if elevator is idle with orders in stopArray
     	if stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down {
	        elev_cmd_chan <- config.FloorReached
          delete_order_chan <- setFloorFalse()
	        elev_cmd_chan <- config.Finished
	    } else if !isEmpty(stopArray, elevio.Ground_floor, config.Current_floor){
        	elev_cmd_chan <- config.GoDown
      } else if !isEmpty(stopArray, config.Current_floor+1, elevio.Num_floors){
        	elev_cmd_chan <- config.GoUp
      } else {
        	elev_cmd_chan <- config.Finished
      }
    }
  }
}
