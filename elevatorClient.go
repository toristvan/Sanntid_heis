package main

import (
    "./driverModule/elevio"
    "./queueModule"
    "./fsmModule"
    ."fmt"
	  "time"
)

const localID = 1

type executeOrderStruct struct{
  active bool
  Floor int
  Button elevio.ButtonType
}

var executeOrderQueue[4] executeOrderStruct
var index = 1

func dummyCostFunc(hallCall elevio.ButtonType, floor int) int {
  return 1
}


func executeOrder(order_chan <-chan queue.OrderStruct, pending_orders chan<- executeOrderStruct){
  select{
  case new_order := <- order_chan:   //Input from queue

      executeOrderQueue[index].Floor  = new_order.Floor
      executeOrderQueue[index].Button = new_order.Button
      executeOrderQueue[index].active = true

      Println("executeOrderQueue", executeOrderQueue[new_order.Floor].Floor)
      pending_orders <- executeOrderQueue[index]
  }
}

func main(){

    var current_order executeOrderStruct
    var current_floor int

    var elevEvent elevio.ButtonEvent
    next_floor := elevEvent.Floor
    order_type := elevEvent.Button


    new_command         := make(chan fsm.ElevCommand)
    status_elev_state   := make(chan fsm.Status)
    sync_elev_state     := make(chan fsm.Status)

    drv_buttons := make(chan elevio.ButtonEvent)
    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)

    order_chan     := make(chan queue.OrderStruct)
    input_queue    := make(chan queue.OrderStruct)
    pending_orders := make(chan executeOrderStruct, 5)

    current_floor = fsm.ElevatorInit()
    queue.InitQueue()

    go elevio.PollButtons(drv_buttons)
    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
	  go elevio.PollStopButton(drv_stop)
	  //go queue.Queue(input_chan, order_chan)

    for {
      go fsm.ElevStateMachine(status_elev_state, sync_elev_state, order_type, next_floor)
      go fsm.ElevInputCommand(new_command)
      go queue.Queue(input_queue, order_chan)
      go executeOrder(order_chan, pending_orders)

      select {
      case button_input := <-drv_buttons:
        var new_order queue.OrderStruct
          //sende ordre til andre her
    		new_order.Button     = button_input.Button
    		new_order.Floor      = button_input.Floor
        new_order.ElevID     = dummyCostFunc(button_input.Button, button_input.Floor)
        new_order.Timestamp  = time.Now()
    		input_queue <- new_order

      case execute_order := <- pending_orders:
        next_floor    = execute_order.Floor
        order_type    = execute_order.Button

        current_order = execute_order

        //Add to watchdog here

  			 elevio.SetButtonLamp(order_type, next_floor, true)

         switch order_type {
         case elevio.BT_HallUp, elevio.BT_HallDown:
            Println("Hall call")
            Println("Floor:", current_order.Floor)

            if next_floor < current_floor{
              new_command <- fsm.GoDown
            } else if next_floor > current_floor {
              new_command <- fsm.GoUp
            } else {
              new_command <- fsm.FloorReached
            }
         case elevio.BT_Cab:
            Println("Cab call")
            Println("Floor:", current_order.Floor)

            if next_floor < current_floor{
              new_command <- fsm.GoDown
            } else if next_floor > current_floor {
              new_command <- fsm.GoUp
            } else {
              new_command <- fsm.FloorReached
            }
          }

          sync_elev_state <- fsm.Active

      case floor_input := <- drv_floors:

          current_floor = floor_input
          elevio.SetFloorIndicator(current_floor)
          Println("Current floor:", current_floor)

          if current_floor == current_order.Floor {
              elevio.SetButtonLamp(current_order.Button, current_floor, false)
                //queue.RemoveOrder(current_floor, localID)
              new_command <- fsm.FloorReached
          }
          sync_elev_state <- fsm.Active

      case current_status := <- status_elev_state:

          switch current_status {
          case fsm.Pending:
              sync_elev_state <- fsm.Done
          case fsm.Active:
              sync_elev_state <- fsm.Pending
          case fsm.Done:
              sync_elev_state <- fsm.Active
          }

      }

    }

}
