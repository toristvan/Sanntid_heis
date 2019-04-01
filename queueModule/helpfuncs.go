package queue

import (
    "../functionalityModule"
    "../driverModule/elevio"
    "../elevsmModule"
    "../configPackage"
    "time"
)


var orderQueue [config.Num_elevs][config.Queue_size] config.OrderStruct

func InitQueue(){
    for j := 0; j< config.Num_elevs; j++{
        for i := 0; i< config.Queue_size; i++{
            orderQueue[j][i] = invalidateOrder(orderQueue[j][i]) 
            orderQueue[j][i].ElevID = j
        }
    }
}

func invalidateOrder(order config.OrderStruct) config.OrderStruct {
    order.Button = config.BT_HallUp
    order.Floor = -1
    order.Cost = config.Max_cost
    order.Cmd = config.OrdrInv
    order.MasterID = -1
    order.SenderID = -1
    return order
}

//Lower cost = better suited for order
func GenericCostFunction(order config.OrderStruct) int {
    var cost int
    var distance int = elevopr.GetCurrentFloor() - order.Floor
    var abs_distance int
    if distance < 0 {
        abs_distance = -distance
    }else{
        abs_distance = distance
    }
    switch elevsm.RetrieveElevState(){
    case config.Idle:
        switch distance == 0{
        case true:
            cost = abs_distance - 3 - config.Num_floors
        case false:
            cost =  abs_distance - 1 - config.Num_floors 
        }
    case config.AtFloor:
        cost =  abs_distance - config.Num_floors
    case config.GoingUp:
        switch distance < 0{ 
        case true:
            cost =  abs_distance
        case false:
            cost = abs_distance - 2 - config.Num_floors
        }
    case config.GoingDown:
        switch distance < 0{
        case true:
            cost = abs_distance - 2 - config.Num_floors
        case false:
        cost = abs_distance 
        }
    }
    return cost
}

func RetrieveQueue() [config.Num_elevs][config.Queue_size]config.OrderStruct{
    return orderQueue
}

func insertOrder(order config.OrderStruct, index int){
    for i := config.Queue_size - 1; i > index; i--{
        orderQueue[order.ElevID][i] = orderQueue[order.ElevID][i-1]
    }
    order.Timestamp = time.Now()
    orderQueue[order.ElevID][index] = order
}

func addToQueue(order config.OrderStruct, set_lights bool) {
    current_state := elevsm.RetrieveElevState()
    if orderQueue[order.ElevID][0].Floor == -1{
        insertOrder(order, 0)
    } else if current_state == config.GoingUp && order.ElevID == config.Local_ID{
        if order.Floor < orderQueue[order.ElevID][0].Floor {
            insertOrder(order, 0)
        }
    } else if current_state == config.GoingDown && order.ElevID == config.Local_ID{
        if order.Floor > orderQueue[order.ElevID][0].Floor {
            insertOrder(order, 0)
        }
    } else {
        for i := 0; i < config.Queue_size; i++{
            if orderQueue[order.ElevID][i].Floor == -1 {
                insertOrder(order, i)
                break
            }
        }
    }
    //Sets designated lights based on bool passed as input. 
    //Fault tolerance reasons; don't set light if not sure of completion
    if set_lights && !(order.Button == config.BT_Cab && order.ElevID != config.Local_ID){
        elevio.SetButtonLamp(order.Button, order.Floor, true)
    }
}

func RemoveOrder(floor int, id int){
    for i := 0; i < config.Num_elevs; i++ {
        for j := 0; j < config.Queue_size; j++{
            if  orderQueue[i][j].Floor == floor && (orderQueue[i][j].Button != config.BT_Cab || id == i) { 
                orderQueue[i][j] = invalidateOrder(orderQueue[i][j])
            }
        }
    }
    if id == config.Local_ID {
        elevio.SetButtonLamp(config.ButtonType(config.BT_Cab), floor, false)
    } 
    for i := config.BT_HallUp; i < config.BT_Cab  ; i++{
        elevio.SetButtonLamp(config.ButtonType(i), floor, false) 
    }
}

func inQueue(order config.OrderStruct) bool{
    if order.Button != config.BT_Cab{
        for i := 0; i < config.Num_elevs; i++ {
            for j := 0; j < config.Queue_size; j++ {
                if order.Floor == orderQueue[i][j].Floor && order.Button == orderQueue[i][j].Button  {
                    return true
                }
            }
        }
    }else{
        for j := 0; j < config.Queue_size; j++ {
            if order.Floor == orderQueue[config.Local_ID][j].Floor  && order.Button == orderQueue[config.Local_ID][j].Button{
                return true
            }
        }           
    }
    return false
}