package elevopr

import (
    "../queueModule"
    "../elevsmModule"
    "../configPackage"
    "time"
)

type floorStatus struct{
    stop_up bool
    stop_down bool
}

//What floors to stop at
var stopArray[config.Num_floors] floorStatus

func initStopArray(){
    for i := config.Ground_floor ; i < config.Num_floors ; i++ {
        stopArray[i].stop_up = false
        stopArray[i].stop_down = false
    }
}

//If no more stops scheduled
func isEmpty(arr [config.Num_floors]floorStatus, from int, to int) bool{
    for i := from ; i < to ; i++{
        if arr[i].stop_up || arr[i].stop_down {
            return false
        }
    }
    return true
}

//Sends alert when elevator is idle and stopArray is not empty
func ElevWakeUp(wakeup_chan chan<- bool){
    for {
        <-time.After(1*time.Second)
        if !isEmpty(stopArray, config.Ground_floor, config.Num_floors) && (elevsm.RetrieveElevState() == config.Idle) {
            wakeup_chan <- true
        }
    }
}

//Schedules stop in stopArray
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

//Converts buttonpress to orders
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

//Removes scheduled stop from stopArray. Returns an order with delete command
func setFloorFalse() config.OrderStruct{
    var order_to_delete config.OrderStruct
    order_to_delete.Floor = config.Current_floor
    order_to_delete.ElevID = config.Local_ID
    order_to_delete.Cmd = config.OrdrDelete

    stopArray[config.Current_floor].stop_up = false
    stopArray[config.Current_floor].stop_down = false
    queue.RemoveOrder(config.Current_floor, config.Local_ID)
    return order_to_delete
}
