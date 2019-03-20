package main

import (
    "./driverModule/elevio"
    ."fmt"
		"time"
)

type state int
const (
	idle 			state = 0
	goingUp 	state = 1
	goingDown state = 2
	atFloor 	state = 4
)
type event int
const (
	newOrder 			event = 0
	goUp		 			event = 1
	goDown 	 			event = 2
	floorReached 	event = 3
	finished 			event = 4
  wait          event = 5
)

var State state
var new_Event event

func initElevToKnownFloor() int {
  var floor int
  var done bool
  drv_floors  := make(chan int)

  go elevio.PollFloorSensor(drv_floors)

  for{
    select{
    case sensor := <- drv_floors:
        floor = sensor
        done = true
    default:
      elevio.SetMotorDirection(elevio.MD_Down)
    }
    if done {
			elevio.SetMotorDirection(elevio.MD_Stop)
      break
    }
  }
  return floor
}

func main(){
    var buttonEvent elevio.ButtonEvent
    floor_Order := buttonEvent.Floor
    button := buttonEvent.Button

    //wait_for_input := make(chan bool)
    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)

		elevio.Init("localhost:15657")
		currentFloor := initElevToKnownFloor()
		State = idle
		Println("Ready")

		go elevio.PollButtons(drv_buttons)
		go elevio.PollFloorSensor(drv_floors)
		go elevio.PollObstructionSwitch(drv_obstr)
		go elevio.PollStopButton(drv_stop)
    //go interruptButtons(drv_buttons, wait_for_input)

    for {
        select {
        case button_Input := <- drv_buttons:
            floor_Order = button_Input.Floor
            button = button_Input.Button
						elevio.SetButtonLamp(button, floor_Order, true)

						switch button {
						case 0, 1:
							Println("Hall call")
							Println("botton input:", button, "floor:", floor_Order)

							if floor_Order < currentFloor{
								new_Event = goDown
							} else if floor_Order > currentFloor {
								new_Event = goUp
							} else {
								new_Event = floorReached
							}

						case 2:
              //Til nå gjør denne casen det samme som hall call
							Println("Cab call")
							Println("botton input:", button, "floor:", floor_Order)

							if floor_Order < currentFloor{
								new_Event = goDown
							} else if floor_Order > currentFloor {
								new_Event = goUp
							} else {
								new_Event = floorReached
							}
						}

        case floor_sensorInput := <- drv_floors:
            currentFloor = floor_sensorInput
            Println("currentFloor:", currentFloor)

						if currentFloor == floor_Order {
							elevio.SetButtonLamp(button, floor_Order, false)
							new_Event = floorReached
						}

				default:

					switch State{
					case idle:
						elevio.SetDoorOpenLamp(false)

						switch new_Event{
						case goDown:
							State = goingDown
						case goUp:
							State = goingUp
						case floorReached:
							State = atFloor
						}

					case goingUp:
						if new_Event == floorReached{
							elevio.SetMotorDirection(elevio.MD_Stop)
							State = atFloor
						} else{
							elevio.SetMotorDirection(elevio.MD_Up)
              time.Sleep(100 * time.Millisecond)
						}

					case goingDown:
						if new_Event == floorReached{
							elevio.SetMotorDirection(elevio.MD_Stop)
							State = atFloor
						} else {
							elevio.SetMotorDirection(elevio.MD_Down)
              time.Sleep(100 * time.Millisecond)
						}

					case atFloor:
						elevio.SetDoorOpenLamp(true)
						elevio.SetButtonLamp(button, floor_Order, false)

            Println("Floor reached")
						time.Sleep(3000 * time.Millisecond)

            new_Event = wait
						State = idle
					}

      }
    }
}
