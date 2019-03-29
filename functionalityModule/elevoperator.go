package elevopr

import (
    "../driverModule/elevio"
    "../elevsmModule"
    "../configPackage"
)

func ElevOperator(elev_cmd_chan chan<- config.ElevCommand, delete_order_chan chan<- config.OrderStruct, wakeup_chan <-chan bool, drv_floors_chan <-chan int){
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
                //If more orders, continue in designated direction
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
                //If more orders, continue in designated direction
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
