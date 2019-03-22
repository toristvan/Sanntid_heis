package fsm

import(
  "./../driverModule/elevio"
  ."fmt"
  "time"
)

// "<-chan" receive
// "chan<-" send

type ElevStateType int
const (
	Idle 				  ElevStateType = 0
	GoingUp 			ElevStateType = 1
	GoingDown 		ElevStateType = 2
	AtFloor 			ElevStateType = 4
)
type ElevCommand int
const (
	NewOrder 			 ElevCommand = 0
	GoUp		 		   ElevCommand = 1
	GoDown 	 			 ElevCommand = 2
	FloorReached 	 ElevCommand = 3
	Finished 			 ElevCommand = 4
  Wait           ElevCommand = 5
)
type Status int
const (
  Pending Status = 0
  Active  Status = 1
  Done    Status = 2
)

var elev_state ElevStateType
var new_command ElevCommand

func ElevatorInit() int {
  var current_floor int

  elevio.Init("localhost:15657")
  elev_state = Idle
  current_floor = elevio.Num_floors + 1
  Println("Ready")

  return current_floor
}

func ElevInputCommand(command <-chan ElevCommand){
  select{
  case  new :=  <-command:
    new_command = new
  }
}

func ElevStateMachine(status_elev_state chan <- Status, sync_elev_state <- chan Status, button elevio.ButtonType, floor int){
      select{
      case sync := <- sync_elev_state:
        time.Sleep(50 * time.Millisecond)
        switch sync{
        case Active:
          status_elev_state <- Pending
        }

        switch elev_state{
        case Idle:
          Println("Idle")
          elevio.SetDoorOpenLamp(false)

          switch new_command{
          case GoDown:
            status_elev_state <- Done
            elev_state = GoingDown
          case GoUp:
            status_elev_state <- Done
            elev_state = GoingUp
          case FloorReached:
            status_elev_state <- Done
            elev_state = AtFloor
          case Wait:
            elev_state = Idle
          }

        case GoingUp:
          if new_command == FloorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            status_elev_state <- Done
            elev_state = AtFloor
          } else{
            elevio.SetMotorDirection(elevio.MD_Up)
            time.Sleep(100 * time.Millisecond)
          }

        case GoingDown:
          if new_command == FloorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            status_elev_state <- Done
            elev_state = AtFloor
          } else {
            elevio.SetMotorDirection(elevio.MD_Down)
            time.Sleep(100 * time.Millisecond)
          }

        case AtFloor:
          elevio.SetDoorOpenLamp(true)
          elevio.SetButtonLamp(button, floor, false)

          Println("Floor reached")

          new_command = Wait
          elev_state = Idle
          status_elev_state <- Done
        }
      }
}
