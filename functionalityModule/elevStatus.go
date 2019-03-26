package elevclient

import (
    "../driverModule/elevio"
    "../fsmModule"
    "../configPackage"
    "fmt"
    "time"
)

//Function checking if elev is 'dead'
func IsElevDead(is_dead_chan chan<- bool){
	drv_floors := make (chan int)
	var dead bool = false
	go elevio.PollFloorSensor(drv_floors)
	for{
		for fsm.RetrieveElevState() == config.GoingUp || fsm.RetrieveElevState() == config.GoingDown {
			deadTimer := time.NewTicker(5*time.Second)
			defer deadTimer.Stop()
			fmt.Println("Waiting for floor....")
			select{
			case <- drv_floors:
				if dead{
					dead = false
					is_dead_chan <- dead
				}
				time.Sleep(1*time.Second)

			case <- deadTimer.C:
				if !dead{
					dead = true
					is_dead_chan <- dead
				}

			}
		}
	}
}