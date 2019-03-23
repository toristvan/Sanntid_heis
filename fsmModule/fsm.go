package fsm

import(
  "./../configPackage"
  "./../driverModule/elevio"
  ."fmt"
  "time"
)

// "<-chan" receive
// "chan<-" send

var elev_state config.ElevStateType
var new_command config.ElevCommand

func ElevatorInit() int {
  var current_floor int

  elevio.Init("localhost:15657")
  elev_state = config.Idle
  current_floor = elevio.Num_floors + 1
  Println("Ready")

  return current_floor
}

func ElevInputCommand(command <-chan config.ElevCommand){
  select{
  case  new :=  <-command:
    new_command = new
  }
}

func RetrieveElevState() config.ElevStateType{     //Any better way to do this?
  return elev_state
}

func ElevStateMachine(status_elev_state chan <- config.Status, sync_elev_state <- chan config.Status, button config.ButtonType, floor int){
      select{
      case sync := <- sync_elev_state:
        time.Sleep(50 * time.Millisecond)
        switch sync{
        case config.Active:
          status_elev_state <- config.Pending
        }

        switch elev_state{
        case config.Idle:
          Println("idle")
          elevio.SetDoorOpenLamp(false)

          switch new_command{
          case config.GoDown:
            status_elev_state <- config.Done
            elev_state = config.GoingDown
          case config.GoUp:
            status_elev_state <- config.Done
            elev_state = config.GoingUp
          case config.FloorReached:
            status_elev_state <- config.Done
            elev_state = config.AtFloor
          case config.Wait:
            elev_state = config.Idle
          }

        case config.GoingUp:
          if new_command == config.FloorReached{
            elevio.SetMotorDirection(config.MD_Stop)
            status_elev_state <- config.Done
            elev_state = config.AtFloor
          } else{
            elevio.SetMotorDirection(config.MD_Up)
            time.Sleep(100 * time.Millisecond)
          }

        case config.GoingDown:
          if new_command == config.FloorReached{
            elevio.SetMotorDirection(config.MD_Stop)
            status_elev_state <- config.Done
            elev_state = config.AtFloor
          } else {
            elevio.SetMotorDirection(config.MD_Down)
            time.Sleep(100 * time.Millisecond)
          }

        case config.AtFloor:
          elevio.SetDoorOpenLamp(true)
          elevio.SetButtonLamp(button, floor, false)

          Println("Floor reached")

          new_command = config.Wait
          elev_state = config.Idle
          status_elev_state <- config.Done
        }
      }
}
