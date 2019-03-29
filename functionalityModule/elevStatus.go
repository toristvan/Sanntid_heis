package elevclient

import (
    //"../driverModule/elevio"
    "../fsmModule"
    "../configPackage"
    //"fmt"
    "time"
)

//Function checking if elev is 'dead'
func IsElevDead(is_dead_chan chan<- bool, drv_floors_chan <-chan int){
	var dead bool = false
	for{
		for fsm.RetrieveElevState() == config.GoingUp || fsm.RetrieveElevState() == config.GoingDown {
			//deadTimer := time.NewTicker(5*time.Second)
			//defer deadTimer.Stop()
			//fmt.Println("Waiting for floor....")
			select{
			case <- drv_floors_chan:
				if dead{
					dead = false
					is_dead_chan <- dead
				}

			case <- time.After(5*time.Second):
				if !dead && fsm.RetrieveElevState() == config.GoingUp || fsm.RetrieveElevState() == config.GoingDown{
					dead = true
					is_dead_chan <- dead
				}

			}
		}
		<-time.After(50*time.Millisecond)
	}
}