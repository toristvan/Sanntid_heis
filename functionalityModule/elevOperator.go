package elevopr

import (
    "../driverModule/elevio"
    "../queueModule"
    "../elevsmModule"
    "../configPackage"
    "time"
)

type floorStatus struct{
  stop_up bool
  stop_down bool
}

//var config.Current_floor int
var stopArray[config.Num_floors] floorStatus

func initStopArray(){
  for i := config.Ground_floor ; i < config.Num_floors ; i++ {
    stopArray[i].stop_up = false
    stopArray[i].stop_down = false
  }
}

func isEmpty(arr [config.Num_floors]floorStatus, from int, to int) bool{
  	for i := from ; i < to ; i++{
    	if arr[i].stop_up || arr[i].stop_down {
      		return false
    	}
  	}
  	return true
}

//An alert function which sends alert when elevator is idle when stopArray is not empty
func ElevWakeUp(wakeup_chan chan<- bool){
  for {
    <-time.After(1*time.Second)
    if !isEmpty(stopArray, config.Ground_floor, config.Num_floors) && (elevsm.RetrieveElevState() == config.Idle) {
      wakeup_chan <- true
    }
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
     // fmt.Printf("Added to stopArray\n") //To check for crash when going idle for long
    }
  }
}

func IOwrapper(distr_order_chan chan<- config.OrderStruct, drv_buttons_chan <-chan config.ButtonEvent){

  for{
      button_input := <-drv_buttons_chan
      var new_order config.OrderStruct
      new_order.ElevID    = config.Local_ID
      new_order.Button    = button_input.Button
      new_order.Floor     = button_input.Floor
      new_order.MasterID  = config.Local_ID

      if (new_order.Button != config.BT_Cab){
        new_order.Cmd = config.CostReq
      } else{
        new_order.Cmd = config.OrdrAdd
      }
      distr_order_chan <- new_order

  }
}

func setFloorFalse() config.OrderStruct{
  var order_to_delete config.OrderStruct

  order_to_delete.Floor = config.Current_floor
  order_to_delete.ElevID = config.Local_ID
  order_to_delete.Cmd = config.OrdrDelete

  stopArray[config.Current_floor].stop_up = false
  stopArray[config.Current_floor].stop_down = false
  queue.RemoveOrder(config.Current_floor, config.Local_ID)
  return order_to_delete

}


func ElevRunner(elev_cmd_chan chan<- config.ElevCommand, delete_order_chan chan<- config.OrderStruct, wakeup_chan <-chan bool, drv_floors_chan <-chan int){
  var current_state config.ElevStateType = config.Idle
  for{
    select{
    case config.Current_floor = <- drv_floors_chan:
      current_state = elevsm.RetrieveElevState()
      elevio.SetFloorIndicator(config.Current_floor)

      switch current_state{
        case config.GoingUp:
          if (stopArray[config.Current_floor].stop_up) || (stopArray[config.Current_floor].stop_down && isEmpty(stopArray, config.Current_floor+1, config.Num_floors)) {
            //Stop routine
            elev_cmd_chan <- config.FloorReached
            delete_order_chan <- setFloorFalse()
        	}
        	//Stop again if new order received when at floor with open door
        	if (stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down) && elevsm.RetrieveElevState() == config.AtFloor {
        	 	elev_cmd_chan <- config.FloorReached
	        	delete_order_chan <- setFloorFalse()
        	}
          //If more orders above current floor, continue upwards
        	if !isEmpty(stopArray, config.Current_floor+1, config.Num_floors){
            	elev_cmd_chan <- config.GoUp
          } else if !isEmpty(stopArray, config.Ground_floor, config.Current_floor){
            	elev_cmd_chan <- config.GoDown
          } else{
            	elev_cmd_chan <- config.Finished
          }

        case config.GoingDown:
        	if (stopArray[config.Current_floor].stop_down) || (stopArray[config.Current_floor].stop_up && isEmpty(stopArray, config.Ground_floor, config.Current_floor)) {
				    //Stop routine
        		elev_cmd_chan <- config.FloorReached
        		delete_order_chan <- setFloorFalse()
        	}
        	//Stop again if new order received when at floor with open door
        	if (stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down) && elevsm.RetrieveElevState() == config.AtFloor {
        		elev_cmd_chan <- config.FloorReached
	        	delete_order_chan <- setFloorFalse()

        	}
          //If more orders below current floor, continue downwards
        	if !isEmpty(stopArray, config.Ground_floor, config.Current_floor){
            	elev_cmd_chan <- config.GoDown
          } else if !isEmpty(stopArray, config.Current_floor+1, config.Num_floors){
            	elev_cmd_chan <- config.GoUp
          } else{
            	elev_cmd_chan <- config.Finished
          }
        }
    //Alert if elevator is idle with orders in stopArray
    case <- wakeup_chan:       
     	if stopArray[config.Current_floor].stop_up || stopArray[config.Current_floor].stop_down {
	        elev_cmd_chan <- config.FloorReached
          delete_order_chan <- setFloorFalse()
	        elev_cmd_chan <- config.Finished
	    } else if !isEmpty(stopArray, config.Ground_floor, config.Current_floor){
        	elev_cmd_chan <- config.GoDown
      } else if !isEmpty(stopArray, config.Current_floor+1, config.Num_floors){
        	elev_cmd_chan <- config.GoUp
      } else {
        	elev_cmd_chan <- config.Finished
      }
    }
  }
}
