package main

import (
    "./driverModule/elevio"
    "./queueModule"
    ."fmt"
	  "time"
)

type ElevStateType int
const (
	idle 				  ElevStateType = 0
	goingUp 			ElevStateType = 1
	goingDown 		ElevStateType = 2
	atFloor 			ElevStateType = 4
)
type elevCommand int
const (
	newOrder 			 elevCommand = 0
	goUp		 		   elevCommand = 1
	goDown 	 			 elevCommand = 2
	floorReached 	 elevCommand = 3
	finished 			 elevCommand = 4
  wait            elevCommand = 5
)
type status int
const (
  pending status = 0
  active  status = 1
  done    status = 2
)

const localID = 0
var elev_state ElevStateType
var new_command elevCommand

func fsmElevator(job_status chan <- status, job_sync <-chan status ,button elevio.ButtonType, floor int){
      select{
      case sync := <- job_sync:
        switch sync{
        case active:
          job_status <- pending
        }

        switch elev_state{
        case idle:
          Println("idle")
          elevio.SetDoorOpenLamp(false)

          switch new_command{
          case goDown:
            job_status <- done
            elev_state = goingDown
          case goUp:
            job_status <- done
            elev_state = goingUp
          case floorReached:
            job_status <- done
            elev_state = atFloor
          case wait:
            elev_state = idle
          }

        case goingUp:
          Println("goingUp")
          if new_command == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            elev_state = atFloor
          } else{
            elevio.SetMotorDirection(elevio.MD_Up)
            time.Sleep(100 * time.Millisecond)
          }

        case goingDown:
          Println("goingDown")
          if new_command == floorReached{
            elevio.SetMotorDirection(elevio.MD_Stop)
            job_status <- done
            elev_state = atFloor
          } else {
            elevio.SetMotorDirection(elevio.MD_Down)
            time.Sleep(100 * time.Millisecond)
          }

        case atFloor:
          elevio.SetDoorOpenLamp(true)
          elevio.SetButtonLamp(button, floor, false)

          Println("Floor reached")

          new_command = wait
          elev_state = idle
          job_status <- done
        }
      }
}

func main(){
    var current_order queue.OrderStruct
    var current_floor int

    var elevEvent elevio.ButtonEvent
    next_floor := elevEvent.Floor
    order_type := elevEvent.Button

    job_status   := make(chan status)
    job_sync     := make(chan status)

    drv_floors  := make(chan int)
    drv_obstr   := make(chan bool)
    drv_stop    := make(chan bool)
    order_chan := make(chan queue.OrderStruct)

    //defer close(drv_floors)
    //defer close(drv_obstr)
    //defer close(drv_stop)
    //defer close(order_chan)

    elevio.Init("localhost:15657")
	  elev_state = idle
	  current_floor = elevio.Num_floors + 1
    Println("Ready")
	  Println("Ready")
	  Println("Current floor:", current_floor)

    go elevio.PollFloorSensor(drv_floors)
    go elevio.PollObstructionSwitch(drv_obstr)
	  go elevio.PollStopButton(drv_stop)
	  go queue.Queue(order_chan)


    for {
         go fsmElevator(job_status, job_sync, order_type, next_floor)

         select {
         case new_order := <- order_chan:   //Input from queue
              next_floor = new_order.Floor
              order_type = new_order.Button
              current_order = new_order

              //Add to watchdog here

  			      elevio.SetButtonLamp(order_type, next_floor, true)

              switch order_type {
              case elevio.BT_HallUp, elevio.BT_HallDown:
                Println("Hall call")
                Println("Floor:", current_order.Floor)

                if next_floor < current_floor{
                  new_command = goDown
                } else if next_floor > current_floor {
                  new_command = goUp
                } else {
                  new_command = floorReached
                }

              case elevio.BT_Cab:
                Println("Cab call")
                Println("Floor:", current_order.Floor)

                if next_floor < current_floor{
                  new_command = goDown
                } else if next_floor > current_floor {
                  new_command = goUp
                } else {
                  new_command = floorReached
                }
              }

              job_sync <- active

         case floor_input := <- drv_floors:
              current_floor = floor_input
              elevio.SetFloorIndicator(current_floor)
              Println("Current floor:", current_floor)

              if current_floor == current_order.Floor {
                elevio.SetButtonLamp(current_order.Button, current_floor, false)
                queue.RemoveOrder(current_floor, localID)
                new_command = floorReached
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
