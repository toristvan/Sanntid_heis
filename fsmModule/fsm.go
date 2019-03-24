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
  current_floor = elevio.Num_floors + 1 //Why?
  //current_floor = 
  Println("Ready")

  return current_floor
}
// Move to elev state machine
func ElevInputCommand(command <-chan config.ElevCommand){
  select{
  case new_cmd :=  <-command:
    new_command = new_cmd
  }
}


func RetrieveElevState() config.ElevStateType{     //Any better way to do this?
  return elev_state
}

func ElevStateMachine(status_elev_state chan <- config.Status, sync_elev_state <- chan config.Status, button config.ButtonType, floor int){
  for{

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
}

func ElevStateMachine2(new_command_chan <-chan config.ElevCommand, current_state *config.ElevStateType){
  for{
      select{
      case new_cmd := <-new_command_chan:
        
        switch new_cmd{
        
        case config.GoUp:
          elevio.SetMotorDirection(config.MD_Up)
          *current_state = config.GoingUp
        case config.GoDown:
          elevio.SetMotorDirection(config.MD_Down)
          *current_state = config.GoingDown

        case config.FloorReached:
          elevio.SetMotorDirection(config.MD_Stop)
          *current_state = config.AtFloor
          elevio.SetDoorOpenLamp(true)
          time.Sleep(3000*time.Millisecond)
          elevio.SetDoorOpenLamp(false)
          //current_state = config.Idle

        case config.Finished:
          *current_state = config.Idle
         
        case config.Wait:

        }

    }

  }
}