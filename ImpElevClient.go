<<<<<<< HEAD
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
	wait			    event = 0
	goUp		 			event = 1
	goDown 	 			event = 2
	floorReached 	event = 3
	finished 			event = 4
  newOrder      event = 5
)

type status int
const (
  pending status = 0
  active  status = 1
  done    status = 2
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

func fsmElevator(job_status chan <- status, job_sync <-chan status ,button elevio.ButtonType, floor int){
      select{
      case sync := <- job_sync:
        switch sync{
        case active:
          job_status <- pending
        }

        switch State{
        case idle:
          Println("idle")
          elevio.SetDoorOpenLamp(false)

          switch new_Event{
          case goDown:
            job_status <- done
            State = goingDown
          case goUp:
            job_status <- done
            State = goingUp
          case floorReached:
            job_status <- done
            State = atFloor
          case wait:
            State = idle
          }

        case goingUp:
          Println("goingUp")
          if new_Event == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            State = atFloor
          } else{
            elevio.SetMotorDirection(elevio.MD_Up)
            time.Sleep(100 * time.Millisecond)
          }

        case goingDown:
          Println("goingDown")
          if new_Event == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            State = atFloor
          } else {
            elevio.SetMotorDirection(elevio.MD_Down)
            time.Sleep(100 * time.Millisecond)
          }

        case atFloor:
          elevio.SetDoorOpenLamp(true)
          elevio.SetButtonLamp(button, floor, false)

          Println("Floor reached")

          new_Event = wait
          State = idle
          job_status <- done
        }
      }
}

func main(){
    var buttonEvent elevio.ButtonEvent
    floor_Order := buttonEvent.Floor
    button := buttonEvent.Button

    job_status   := make(chan status)
    job_sync     := make(chan status)

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

    for {
        go fsmElevator(job_status, job_sync, button, floor_Order)

        select {
        case button_Input := <- drv_buttons:
            floor_Order = button_Input.Floor
            button = button_Input.Button

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

            elevio.SetButtonLamp(button, floor_Order, true)
            job_sync <- active

        case floor_sensorInput := <- drv_floors:
            currentFloor = floor_sensorInput
            Println("currentFloor:", currentFloor)

						if currentFloor == floor_Order {
							elevio.SetButtonLamp(button, floor_Order, false)
							new_Event = floorReached
						}
            job_sync <- active

        case rec_status := <- job_status:
          switch rec_status {
          case pending:
            job_sync <- done
          case active:
            job_sync <- pending
          case done:
            job_sync <- active
          }

      }
    }
}
=======
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
	wait			    event = 0
	goUp		 			event = 1
	goDown 	 			event = 2
	floorReached 	event = 3
	finished 			event = 4
  newOrder      event = 5
)

type status int
const (
  pending status = 0
  active  status = 1
  done    status = 2
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

func fsmElevator(job_status chan <- status, job_sync <-chan status ,button elevio.ButtonType, floor int){
      select{
      case sync := <- job_sync:
        switch sync{
        case active:
          job_status <- pending
        }

        switch State{
        case idle:
          Println("idle")
          elevio.SetDoorOpenLamp(false)

          switch new_Event{
          case goDown:
            job_status <- done
            State = goingDown
          case goUp:
            job_status <- done
            State = goingUp
          case floorReached:
            job_status <- done
            State = atFloor
          case wait:
            State = idle
          }

        case goingUp:
          Println("goingUp")
          if new_Event == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            State = atFloor
          } else{
            elevio.SetMotorDirection(elevio.MD_Up)
            time.Sleep(100 * time.Millisecond)
          }

        case goingDown:
          Println("goingDown")
          if new_Event == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            State = atFloor
          } else {
            elevio.SetMotorDirection(elevio.MD_Down)
            time.Sleep(100 * time.Millisecond)
          }

        case atFloor:
          elevio.SetDoorOpenLamp(true)
          elevio.SetButtonLamp(button, floor, false)

          Println("Floor reached")

          new_Event = wait
          State = idle
          job_status <- done
        }
      }
}

func main(){
    var buttonEvent elevio.ButtonEvent
    floor_Order := buttonEvent.Floor
    button := buttonEvent.Button

    job_status   := make(chan status)
    job_sync     := make(chan status)

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

    for {
        go fsmElevator(job_status, job_sync, button, floor_Order)

        select {
        case button_Input := <- drv_buttons:
            floor_Order = button_Input.Floor
            button = button_Input.Button

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

            elevio.SetButtonLamp(button, floor_Order, true)
            job_sync <- active

        case floor_sensorInput := <- drv_floors:
            currentFloor = floor_sensorInput
            Println("currentFloor:", currentFloor)

						if currentFloor == floor_Order {
							elevio.SetButtonLamp(button, floor_Order, false)
							new_Event = floorReached
						}
            job_sync <- active

        case rec_status := <- job_status:
          switch rec_status {
          case pending:
            job_sync <- done
          case active:
            job_sync <- pending
          case done:
            job_sync <- active
          }

      }
    }
}
>>>>>>> 06afbbbcb96b17b8ee06d1e7857ac7420e50c7ef
