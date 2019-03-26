package elevclient

import (
    //"../queueModule"
    "../driverModule/elevio"
    "../fsmModule"
    "../configPackage"
    //"fmt"
    "time"
)


func IsElevDead(is_dead_chan chan<- bool){
	drv_floors := make (chan bool)
	var dead bool = false
	//go elevio.PollFloorSensor(drv_floors)
	go elevio.PollStopButton(drv_floors)
	//deadTimer := time.NewTimer(5*Time.Second)
	for{
		for fsm.RetrieveElevState() == config.GoingUp || fsm.RetrieveElevState() == config.GoingDown {
			deadTimer := time.NewTicker(2*time.Second)
			defer deadTimer.Stop()
		
			select{
			case <- drv_floors:
				if dead{
					dead = false
					is_dead_chan <- dead
				}

			case <- deadTimer.C:
				if !dead{
					dead = true
					is_dead_chan <- true
				}

			}
		}
	}
}