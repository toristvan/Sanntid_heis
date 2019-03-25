package fsm

import(
  "../configPackage"
  "../driverModule/elevio"
  "fmt"
  "time"
)

// "<-chan" receive
// "chan<-" send

var elev_state config.ElevStateType
var new_command config.ElevCommand

//elev_state is no longer updated? Should use current_state instead?
func RetrieveElevState() config.ElevStateType{     //Any better way to do this?
  return elev_state
}

func ElevStateMachine2(new_command_chan <-chan config.ElevCommand){
  elev_state = config.Idle
  for{
      select{
      case new_cmd := <-new_command_chan:
        fmt.Println("new command")
        switch new_cmd{

        case config.GoUp:
          elevio.SetMotorDirection(config.MD_Up)
          elev_state = config.GoingUp            //fjern hvis unødvendig, hør med Tor
          fmt.Println("Going up")
        case config.GoDown:
          elevio.SetMotorDirection(config.MD_Down)
          elev_state = config.GoingDown            //fjern hvis unødvendig, hør med Tor
          fmt.Println("Going down")
        case config.FloorReached:
          elevio.SetMotorDirection(config.MD_Stop)
          elev_state = config.AtFloor           //fjern hvis unødvendig, hør med Tor
          fmt.Println("At floor")
          elevio.SetDoorOpenLamp(true)
          time.Sleep(3000*time.Millisecond)
          elevio.SetDoorOpenLamp(false)
          //elev_state = config.Idle

        case config.Finished:
          elev_state = config.Idle          //fjern hvis unødvendig, hør med Tor
          fmt.Println("Idle")
        case config.Wait:

        }
    }

  }
}


/*
func ElevStateMachine(execute_order <-chan config.OrderStruct, internal_floor_chan chan int){

  var current_order config.OrderStruct
  var elevEvent config.ButtonEvent
  next_floor := elevEvent.Floor
	order_type := elevEvent.Button

  internal_command_chan        := make(chan config.ElevCommand)
  internal_state_status_chan   := make(chan config.Status)
  internal_state_sync_chan     := make(chan config.Status)

  go ElevSelectState(internal_state_status_chan, internal_state_sync_chan, order_type, next_floor)
  go ElevInputCommand(internal_command_chan)

  for{
    

    select{
    case exe_ord := <- execute_order:
      next_floor    = exe_ord.Floor
      order_type    = exe_ord.Button
      current_order = exe_ord

      elevio.SetButtonLamp(order_type, next_floor, true)
      internal_state_sync_chan <- config.Active

        switch order_type {
        case config.BT_HallUp, config.BT_HallDown:
         Println("Hall call")
         Println("Floor:", current_order.Floor)
         if next_floor < current_floor{
           internal_command_chan <- config.GoDown
         } else if next_floor > current_floor {
           internal_command_chan <- config.GoUp
         } else {
           internal_command_chan <- config.FloorReached
         }

       case config.BT_Cab:
         Println("Cab call")
         Println("Floor:", current_order.Floor)
         if next_floor < current_floor{
           internal_command_chan <- config.GoDown
         } else if next_floor > current_floor {
           internal_command_chan <- config.GoUp
         } else {
           internal_command_chan <- config.FloorReached
         }
       }

     case floor_input := <- internal_floor_chan:
       Println(floor_input)
       current_floor = floor_input
       elevio.SetFloorIndicator(current_floor)
       if current_floor == current_order.Floor {
         elevio.SetButtonLamp(current_order.Button, current_floor, false)
         internal_command_chan <- config.FloorReached
       }
       internal_state_sync_chan <- config.Active

      case current_status := <- internal_state_status_chan:
        switch current_status {
        case config.Pending:
           internal_state_sync_chan <- config.Done
        case config.Active:
           internal_state_sync_chan <- config.Pending
        case config.Done:
           internal_state_sync_chan <- config.Active
        }

    }
  }
}
*/
// Move to elev state machine
/*
func ElevInputCommand(internal_command_chan <-chan config.ElevCommand){
  select{
  case new_cmd :=  <-command:
    new_command = new_cmd
  }
}
*/

/*
func ElevSelectState(status_elev_state chan <- config.Status, sync_elev_state <- chan config.Status, button config.ButtonType, floor int){

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
*/
