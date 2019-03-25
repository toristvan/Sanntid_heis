package elevclient

import (
    "../driverModule/elevio"
    "../queueModule"
    "../fsmModule"
    "../configPackage"
    "fmt"
    "time"
)
/*
=========================== Bothause! ================================
Denne fungerer med å bruke "Test queue" i queueModule (kanskje, har blitt modifisert)

// TODO:
Legge inn retransmit order hvis en blir passert. Line 142->
*/

type floorStatus struct{
  stop bool
  Button config.ButtonType
}

var current_floor int
//convert to int
var stopArray[elevio.Num_floors] floorStatus

func initStopArray(){
  for i := elevio.Ground_floor ; i < elevio.Num_floors ; i++ {
    stopArray[i].stop = false
    stopArray[i].Button = config.BT_Cab
  }
} 

//var index = 1

func isEmpty(arr [elevio.Num_floors]floorStatus, from int, to int) bool{
  	for i := from ; i < to ; i++{
    	if arr[i].stop {
      	return false
    	}
  	}
  	return true
}

//An alert function which sends alert when elevator is idle when stopArray is not empty
func queueAlert(alert_chan chan<- bool){
  for {
    if !isEmpty(stopArray, elevio.Ground_floor, elevio.Num_floors) && (fsm.RetrieveElevState() == config.Idle || fsm.RetrieveElevState() == config.AtFloor){
      fmt.Println("Queue Alert!")
      alert_chan <- true
    }
    time.Sleep(1*time.Second)   //Unload CPU
  }
}

func dummyCostFunc(hallCall config.OrderStruct) int {
  return 1
}

func executeOrder(execute_chan <-chan config.OrderStruct){ //, pending_orders chan<- floorStatus){
  for {
    //select{
    //case 
    new_order := <- execute_chan   //Input from queue

    stopArray[new_order.Floor].Button = new_order.Button
    stopArray[new_order.Floor].stop = true
    fmt.Println(stopArray)
    queue.RemoveOrder(new_order.Floor, config.LocalID)   //Remove order when arriving instead of here

  //pending_orders <- stopArray[index]
  }
}

func IOwrapper(new_order_chan chan<- config.OrderStruct, floor_chan chan<- int){

  drv_floors  := make(chan int)
  drv_buttons := make(chan config.ButtonEvent)

  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollButtons(drv_buttons)

  elevio.SetMotorDirection(config.MD_Down)
  current_floor = <- drv_floors
  elevio.SetMotorDirection(config.MD_Stop)
  elevio.SetFloorIndicator(current_floor)

  for{
      select{
      case button_input := <-drv_buttons:
        var new_order config.OrderStruct
        new_order.ElevID    = config.LocalID
        new_order.Button    = button_input.Button
        new_order.Floor     = button_input.Floor

        if (new_order.Button != config.BT_Cab){
          new_order.Cmd = config.CostReq
        } else{
          new_order.Cmd = config.OrdrAdd
        }

        new_order_chan <- new_order

    case current_floor = <- drv_floors: //kanskje unøvendig, ikke helt sikker
        floor_chan <- current_floor
    }
  }
}

func ElevRunner(){
  //var current_dir config.MotorDirection
  //var prev_dir config.MotorDirection
  //var current_state config.ElevStateType = config.Idle
  //fsm.RetrieveState()
  var prev_state config.ElevStateType = config.Idle


  floor_chan  := make (chan int)
  input_queue := make (chan config.OrderStruct)
  execute_chan := make (chan config.OrderStruct)
  alert_chan := make(chan bool)
  elev_cmd_chan := make (chan config.ElevCommand)

  go executeOrder(execute_chan)
  go queue.Queue(input_queue, execute_chan)
  go IOwrapper(input_queue, floor_chan)
  go queueAlert(alert_chan)
  go fsm.ElevStateMachine2(elev_cmd_chan)

  for{
    select{
    case new_floor := <- floor_chan:
      prev_state = fsm.RetrieveElevState()      
      current_floor = new_floor
      if stopArray[new_floor].stop{
        //Stop routine
        elev_cmd_chan <- config.FloorReached
        stopArray[new_floor].stop = false
        for i := 0; i < 3; i++ {
          elevio.SetButtonLamp(config.ButtonType(i), new_floor, false)  //Switch off all lights associated with floor
        }
        fmt.Println(stopArray)
      }
      elevio.SetFloorIndicator(current_floor)
      
      switch prev_state{
        case config.GoingUp:
          if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
            fmt.Println(prev_state)
            elev_cmd_chan <- config.GoUp
          } else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
            elev_cmd_chan <- config.GoDown
          } else{
            elev_cmd_chan <- config.Finished
          }
        case config.GoingDown:
          if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
            elev_cmd_chan <- config.GoDown
          } else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
            elev_cmd_chan <- config.GoUp
          } else{
            elev_cmd_chan <- config.Finished
          }
        }
        

    case <- alert_chan:       //Channel dedicated to alert if elevator is idle with orders in stopArray
     	if stopArray[current_floor].stop {
	        elev_cmd_chan <- config.FloorReached
	        fmt.Println("3")
	        stopArray[current_floor].stop = false
	        //Turn off lights should be in RemoveArray instead
	        for i := 0; i < 3; i++ {
	          elevio.SetButtonLamp(config.ButtonType(i), current_floor, false)  //Switch off all lights associated with floor
	        }
	        elev_cmd_chan <- config.Finished
	    } else if !isEmpty(stopArray, elevio.Ground_floor, current_floor){
        	elev_cmd_chan <- config.GoDown
        	fmt.Println("1")
      	} else if !isEmpty(stopArray, current_floor+1, elevio.Num_floors){
        	elev_cmd_chan <- config.GoUp
        	fmt.Println("2")
      	} else {
      		fmt.Println("4")
        	elev_cmd_chan <- config.Finished
      	}
    }
  }
}