package IO

import(
  "../driverModule/elevio"
  "../configPackage"
)

/*
Bruke tanken: lage funksjoner som gjør en ting og en ting bra. Blir to ting i dette tilfellet
Distrubiering av knappetrykk fra en plass og setter dette sammen til en OrderStruct
*/

func IOwrapper(internal_new_order_chan chan<- config.OrderStruct, internal_floor_chan chan<- int){
  var new_order config.OrderStruct

  drv_floors  := make(chan int)
  drv_buttons := make(chan config.ButtonEvent)

  go elevio.PollFloorSensor(drv_floors)
  go elevio.PollButtons(drv_buttons)

  for{
    select{
      case button_input := <-drv_buttons:
        new_order.ElevID    = config.LocalID
  	new_order.Button    = button_input.Button
  	new_order.Floor     = button_input.Floor

        internal_new_order_chan <- new_order

      case floor_input := <- drv_floors: //kanskje unøvendig, ikke helt sikker
        internal_floor_chan <- floor_input
    }
  }
}
