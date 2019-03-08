package main

import "./driverModule/elevio"
import "./queueModule"
import "fmt"


func costFunction(newFloor int, dir int ) { // IN: currentFloor, Direction, (Queues?) , OUT: Cost
	floorDiff := (newFloor - currentFloor)
	cost := floorDiff
	if floorDiff*dir > 0 && dir != 0 {
		cost = floorDiff - 1
	} else if floorDiff*dir < 0 && dir != 0 {
		cost = floorDiff + 1
	} else {
		cost = floorDiff
	}
	//Broadcast result with ID
	return cost
}

func main() {

	numFloors := 4
	//queue.fillQueue()
	nextFloor :=  0

	elevio.Init("localhost:15657", numFloors)

	var d elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(d)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)

	for {
		select {
		case a := <-drv_buttons:			//Receive buttonpress
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			//queue.addHallCall(a.Floor, a.Button)
		case a := <-drv_floors:
			if a.Button != 2 {			// If not Cab call
				costFunction(a.Floor, d)	//Calculate cost function and broadcast order
				//Receive cost function results
				//if self, AddToQueue and AddToWatchdog
				//else, AddToWatchdog
			} else {
				nextFloor = a.Floor		//AddToQueue instead
				//AddToWatchdog
			}
			
			elevio.SetButtonLamp(a.Button, a.Floor, true)	//Set lamp after order is sent to watchdog

		case a := <-drv_floors:		//Receive current floor
			fmt.Printf("%+v\n", a)
			if a > nextFloor {
				d = elevio.MD_Down
			} else if a < nextFloor {
				d = elevio.MD_Up
			} else if a == nextFloor {	//else?
				d = elevio.MD_Stop
				elevio.SetDoorOpenLamp(true)
				elevio.SetButtonLamp(a.Button, a.Floor, false)
				time.Sleep(5* time.Seconds)	//Wait 5 s. Maybe not here?
			}
			if queue.checkStop(a, d){
				elevio.SetMotorDirection(elevio.MD_Stop)
				time.sleep(3000*time.Millisecond)
			}
			queue.removeOrder(a, d)
			elevio.SetMotorDirection(d)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)
			if a {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(d)
			}

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
