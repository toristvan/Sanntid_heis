package elevopr

import (
    "github.com/toristvan/elevator/driverModule/elevio"
    "github.com/toristvan/elevator/elevsmModule"
    "github.com/toristvan/elevator/configPackage"
)

func ElevOperator(elev_cmd_chan chan<- config.ElevCommand, delete_order_chan chan<- config.OrderStruct, wakeup_chan <-chan bool, drv_floors_chan <-chan int){
    var current_state config.ElevStateType = config.Idle
    for{
        select{
        case current_floor = <- drv_floors_chan:
            current_state = elevsm.RetrieveElevState()
            elevio.SetFloorIndicator(current_floor)

            switch current_state{
            case config.GoingUp:
                if (stopArray[current_floor].stop_up) || (stopArray[current_floor].stop_down && isEmpty(stopArray, current_floor+1, config.Num_floors)) {
                    elev_cmd_chan <- config.FloorReached
                    delete_order_chan <- setFloorFalse()
            	}
            	if (stopArray[current_floor].stop_up || stopArray[current_floor].stop_down) && elevsm.RetrieveElevState() == config.AtFloor {
            	 	elev_cmd_chan <- config.FloorReached
    	        	delete_order_chan <- setFloorFalse()
            	}
            	if !isEmpty(stopArray, current_floor+1, config.Num_floors){
                    elev_cmd_chan <- config.GoUp
                } else if !isEmpty(stopArray, config.Ground_floor, current_floor){
                    elev_cmd_chan <- config.GoDown
                } else{
                	elev_cmd_chan <- config.Finished
                }

            case config.GoingDown:
                if (stopArray[current_floor].stop_down) || (stopArray[current_floor].stop_up && isEmpty(stopArray, config.Ground_floor, current_floor)) {
            		elev_cmd_chan <- config.FloorReached
            		delete_order_chan <- setFloorFalse()
            	}
            	if (stopArray[current_floor].stop_up || stopArray[current_floor].stop_down) && elevsm.RetrieveElevState() == config.AtFloor {
            		elev_cmd_chan <- config.FloorReached
    	        	delete_order_chan <- setFloorFalse()
            	}
            	if !isEmpty(stopArray, config.Ground_floor, current_floor){
                	elev_cmd_chan <- config.GoDown
                } else if !isEmpty(stopArray, current_floor+1, config.Num_floors){
                	elev_cmd_chan <- config.GoUp
                } else{
                	elev_cmd_chan <- config.Finished
                }
            }
        case <- wakeup_chan:       
            if stopArray[current_floor].stop_up || stopArray[current_floor].stop_down {
    	        elev_cmd_chan <- config.FloorReached
                delete_order_chan <- setFloorFalse()
    	        elev_cmd_chan <- config.Finished
    	    } else if !isEmpty(stopArray, config.Ground_floor, current_floor){
            	elev_cmd_chan <- config.GoDown
            } else if !isEmpty(stopArray, current_floor+1, config.Num_floors){
            	elev_cmd_chan <- config.GoUp
            } else {
            	elev_cmd_chan <- config.Finished
            }
        }
    }
}
