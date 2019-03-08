package queue

import "./../driverModule/elevio"




const _queueSize int = 10
// dir BT_Cab signalizises cab call
// maybe add floorstop array
type orderStruct struct
{
	dir elevio.ButtonType
	floor int
}


// flytte til main?
var orderQueue [_queueSize]orderStruct
func fillQueue(){
	var invalidOrder orderStruct
	invalidOrder.dir = 0
	invalidOrder.floor = -1
	for i := 0; i< _queueSize; i++{
		orderQueue[i] = invalidOrder
	}
}
//var orderQueue := make([] ,_queueSize )

func addHallCall(floor int, dir elevio.ButtonType){
	var order orderStruct
	order.dir = dir
	order.floor = floor
	for i := 0; i< _queueSize; i++{
		if orderQueue[i].floor != -1{
			orderQueue[i] = order
			break
		}
	}
}

func addCabCall(floor int){
	var order orderStruct
	order.dir = elevio.MD_Stop
	order.floor = floor
	for i := 0; i< _queueSize; i++{
		if orderQueue[i].floor != -1{
			orderQueue[i] = order
			break
		}
	}	
}

func removerOrder(floor int, dir elevio.MotorDirection){
	var but elevio.ButtonType
	if dir == elevio.MD_Up{
		but = elevio.BT_HallUp
	} else if dir == elevio.MD_Stop{
		but = elevio.BT_Cab
	} else if dir == elevio.MD_Down{
		but = elevio.BT_HallDown
	}
	for i := 0; i< _queueSize; i++{
		if orderQueue[i].floor == floor && (orderQueue[i].dir == but|| orderQueue[i].dir == elevio.BT_Cab) {
			orderQueue[i].floor = -1
		}
	}
}


func checkStop(floor int, dir elevio.MotorDirection) bool{	
	var but elevio.ButtonType
	if dir == elevio.MD_Up{
		but = elevio.BT_HallUp
	} else if dir == elevio.MD_Stop{
		but = elevio.BT_Cab
	} else if dir == elevio.MD_Down{
		but = elevio.BT_HallDown
	}
	return orderQueue[0].floor == floor && (orderQueue[0].dir == but || orderQueue[0].dir == elevio.BT_Cab) 
}
