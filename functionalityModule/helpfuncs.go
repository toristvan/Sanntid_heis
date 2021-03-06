package elevopr

import (
    "../elevsmModule"
    "../configPackage"
    "time"
)

type floorStatus struct{
    stop_up bool
    stop_down bool
}

var stopArray[config.Num_floors] floorStatus
var current_floor int

func initStopArray(){
    for i := config.Ground_floor ; i < config.Num_floors ; i++ {
        stopArray[i].stop_up = false
        stopArray[i].stop_down = false
    }
}

func GetCurrentFloor() int {
    return current_floor
}

func isEmpty(arr [config.Num_floors]floorStatus, from int, to int) bool{
    for i := from ; i < to ; i++{
        if arr[i].stop_up || arr[i].stop_down {
            return false
        }
    }
    return true
}

func ElevWakeUp(wakeup_chan chan<- bool){
    for {
        <-time.After(1*time.Second)
        if !isEmpty(stopArray, config.Ground_floor, config.Num_floors) && (elevsm.RetrieveElevState() == config.Idle) {
            wakeup_chan <- true
        }
    }
}

func ExecuteOrder(execute_chan <-chan config.OrderStruct){
    for {
        new_order := <- execute_chan 
        switch new_order.Button{
        case config.BT_HallUp:
            stopArray[new_order.Floor].stop_up = true
        case config.BT_HallDown:
            stopArray[new_order.Floor].stop_down = true
        case config.BT_Cab:
            stopArray[new_order.Floor].stop_up = true
            stopArray[new_order.Floor].stop_down = true
        }
    }
}

func IOwrapper(distr_order_chan chan<- config.OrderStruct, drv_buttons_chan <-chan config.ButtonEvent){
    for{
        button_input := <-drv_buttons_chan
        var new_order config.OrderStruct
        new_order.ElevID    = config.Local_ID
        new_order.Button    = button_input.Button
        new_order.Floor     = button_input.Floor
        new_order.MasterID  = config.Local_ID

        if (new_order.Button != config.BT_Cab){
            new_order.Cmd = config.CostReq
        } else{
            new_order.Cmd = config.OrdrAdd
        }
        distr_order_chan <- new_order
    }
}

func setFloorFalse() config.OrderStruct{
    var order_to_delete config.OrderStruct
    order_to_delete.Floor = current_floor
    order_to_delete.ElevID = config.Local_ID
    order_to_delete.Cmd = config.OrdrDelete

    stopArray[current_floor].stop_up = false
    stopArray[current_floor].stop_down = false
    return order_to_delete
}
