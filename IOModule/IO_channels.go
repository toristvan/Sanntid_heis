package IO

import(
  "../driverModule/elevio"
  "../configPackage"
)


func IOwrapper(new_order_chan chan<- config.OrderStruct, internal_floor_chan chan int){

  drv_floors  := make(chan int)
	drv_buttons := make(chan config.ButtonEvent)

  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollButtons(drv_buttons)

  for{
    select{
      case button_input := <-drv_buttons:
        var new_order config.OrderStruct
        new_order.ElevID 		= config.LocalID
  			new_order.Button    = button_input.Button
  			new_order.Floor     = button_input.Floor

        if (new_order.Button != config.BT_Cab){
          new_order.Cmd = config.CostReq
        } else{
          new_order.Cmd = config.OrdrAdd
        }

        new_order_chan <- new_order

      case floor_input := <- drv_floors: //kanskje unøvendig, ikke helt sikker
        floor_chan <- floor_input
    }
  }
}

