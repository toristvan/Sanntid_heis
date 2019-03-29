package fsm

import(
  "../configPackage"
  "../driverModule/elevio"
  //"fmt"
  "time"
)

var elev_state config.ElevStateType
var new_command config.ElevCommand

//Call it behaviour
func RetrieveElevState() config.ElevStateType{
  return elev_state
}

func ElevStateMachine(new_command_chan <-chan config.ElevCommand){
  elev_state = config.Idle
  for{
    select{
    case new_cmd := <-new_command_chan:
      switch new_cmd{
      case config.GoUp:
        elevio.SetMotorDirection(config.MD_Up)
        elev_state = config.GoingUp           
        //fmt.Println("Going up")
      case config.GoDown:
        elevio.SetMotorDirection(config.MD_Down)
        elev_state = config.GoingDown         
        //fmt.Println("Going down")
      case config.FloorReached:
        elev_state = config.AtFloor           
        elevio.SetMotorDirection(config.MD_Stop)
        //fmt.Println("At floor")
        elevio.SetDoorOpenLamp(true)
        <-time.After(2*time.Second)
        //time.Sleep(2000*time.Millisecond)
        elevio.SetDoorOpenLamp(false)
      case config.Finished:
        elev_state = config.Idle 
        //fmt.Println("Idle")
      }
    }
  }
}
