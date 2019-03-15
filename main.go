package main

import (
	"./driverModule/elevio"
	"./networkModule/bcast"
	"./queueModule"
	"fmt"
	"time"
	)

const localID = 0	//ID should be decided manually upon initialization

func main() {

	queue.InitQueue()
	current_floor := 0

	elevio.Init("localhost:15657") //, num_floors)

	var dir elevio.MotorDirection = elevio.MD_Up
	elevio.SetMotorDirection(dir)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	drv_stop := make(chan bool)
	transmit_order := make(chan queue.OrderStruct)
	receive_order := make(chan queue.OrderStruct)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevio.PollStopButton(drv_stop)
	go bcast.Transmitter(16569,transmit_order)
	go bcast.Receiver(16569, receive_order)
	//go RunElevator()

	for {
		select {
		case btn := <-drv_buttons:			//Receive buttonpress
			new_order = queue.CreateOrder(btn.Floor, btn.Button)    //ID of sender, if needed
			transmit_order <- new_order
			if new_order.btn == elevio.BT_Cab {
				queue.AddOrder(btn, current_floor, localID)
			}

		case flr := <-drv_floors:		//Receive current floor
			fmt.Printf("%+v\n", flr)
			if queue.CheckStop(flr, dir, localID){
				dir = elevio.MD_Stop
				elevio.SetMotorDirection(dir)
				elevio.SetDoorOpenLamp(true)
				for i := 0 ; i<3 ; i++ {
					elevio.SetButtonLamp(elevio.ButtonType(i), current_floor, false)	
				}
				time.Sleep(3* time.Second)	//Wait 5 s. Maybe not here?
				elevio.SetDoorOpenLamp(false)
				queue.RemoveOrder(flr, localID)
			}
			
			

			elevio.SetMotorDirection(dir)

		case obstr := <-drv_obstr:
			fmt.Printf("%+v\n", obstr)
			if obstr {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else {
				elevio.SetMotorDirection(dir)
			}

		case stop := <-drv_stop:
			fmt.Printf("%+v\n", stop)
			elevio.SetDoorOpenLamp(false)                 //Midlertidig?
			for f := 0; f < elevio.Num_floors; f++ {
				for b := elevio.ButtonType(0); b < 3; b++ {
					elevio.SetButtonLamp(b, f, false)
				}
			}
		}
	}
}
