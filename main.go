package main

import (
	"./driverModule/elevio"
	"./queueModule"
	"fmt"
	"time"
	)

func costFunction(newFloor int, currentFloor int, dir elevio.MotorDirection ) int { // IN: currentFloor, Direction, (Queues?) , OUT: Cost
	floorDiff := (newFloor - currentFloor)
	cost := floorDiff
	if floorDiff*int(dir) > 0 && dir != 0 {
		cost = floorDiff - 1
	} else if floorDiff*int(dir) < 0 && dir != 0 {
		cost = floorDiff + 1
	} else {
		cost = floorDiff
	}
	//Broadcast result with ID
	return cost
}

func main() {

	numFloors := 4
	queue.InitQueue()
	queue.OrderQueue[1].Floor =  0
	currentFloor := 0

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
	//go RunElevator()

	for {
		select {
		case a := <-drv_buttons:			//Receive buttonpress
			fmt.Printf("%+v\n", a)
			elevio.SetButtonLamp(a.Button, a.Floor, true)

			if a.Button != 2 {			// If not Cab call
				costFunction(a.Floor, currentFloor, d)	//Calculate cost function and broadcast order
				//Receive cost function results
				queue.AddHallCall(a.Floor, a.Button) //if self, AddToQueue and AddToWatchdog
				//else, AddToWatchdog
			} else {
				queue.AddCabCall(a.Floor)
				//AddToWatchdog
			}
			
			elevio.SetButtonLamp(a.Button, a.Floor, true)	//Set lamp after order is sent to watchdog

		case a := <-drv_floors:		//Receive current floor
			fmt.Printf("%+v\n", a)
			currentFloor = a
			if a > queue.OrderQueue[1].Floor {
				d = elevio.MD_Down
			} else if a < queue.OrderQueue[1].Floor {
				d = elevio.MD_Up
			} else if a == queue.OrderQueue[1].Floor {	//else?
				d = elevio.MD_Stop
				elevio.SetMotorDirection(d)
				elevio.SetDoorOpenLamp(true)
				for i := 0 ; i<3 ; i++ {
					elevio.SetButtonLamp(elevio.ButtonType(i), currentFloor, false)	
				}
				time.Sleep(3* time.Second)	//Wait 5 s. Maybe not here?
				elevio.SetDoorOpenLamp(false)
				queue.RemoveOrder(a, d)
			}
			//if queue.CheckStop(a, d){
			//	time.Sleep(3000*time.Millisecond)
			//}

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
			elevio.SetDoorOpenLamp(false)                 //Midlertidig?
			for f := 0; f < numFloors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
