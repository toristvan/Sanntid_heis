package elevopr

import (
    "../elevsmModule"
    "../configPackage"
    "time"
)

//Checking if elev is non-responsive, based on state indiciating movement, but no new floors are reached
func IsElevDead(is_dead_chan chan<- bool, drv_floors_chan <-chan int){
	var dead bool = false
	for{
		for elevsm.RetrieveElevState() == config.GoingUp || elevsm.RetrieveElevState() == config.GoingDown {
			select{
			case <- drv_floors_chan:
				if dead{
					dead = false
					is_dead_chan <- dead
				}
			case <- time.After(5*time.Second):
				//Needs double check here (same as in for-statement) in case state changes
				if !dead && (elevsm.RetrieveElevState() == config.GoingUp || elevsm.RetrieveElevState() == config.GoingDown){
					dead = true
					is_dead_chan <- dead
				}
			}
		}
		<-time.After(50*time.Millisecond)
	}
}