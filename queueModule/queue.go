package queue

import (
	"./../driverModule/elevio"
	"fmt"
	)



const _queueSize int = 10
const _N_elevs int = 3
// dir BT_Cab signalizises cab call
// maybe add floorstop array
type OrderStruct struct
{
	Dir elevio.ButtonType
	Floor int
}


// flytte til main?
var OrderQueue [_N_elevs][_queueSize] OrderStruct
func InitQueue(){
	var invalidOrder OrderStruct
	invalidOrder.Dir = 0
	invalidOrder.Floor = -1
	for j := 0; j<_N_elevs; j++{
		for i := 0; i< _queueSize; i++{
			OrderQueue[j][i] = invalidOrder
		}
	}
}
//var OrderQueue := make([] ,_queueSize )

func AddHallCall(floor int, dir elevio.ButtonType){
	var order OrderStruct
	order.Dir = dir
	order.Floor = floor
	for i := 0; i< _queueSize; i++{
		if OrderQueue[i].Floor == -1{
			OrderQueue[i] = order
			break
		}
	}
	fmt.Printf("Order added: %+v\n Order queue: %+v\n", order, OrderQueue)
}

func AddCabCall(floor int){
	var order OrderStruct
	order.Dir = elevio.MD_Stop
	order.Floor = floor
	for i := 0; i< _queueSize; i++{
		if OrderQueue[i].Floor == -1{
			OrderQueue[i] = order
			break
		}
	}	
}

func RemoveOrder(floor int, dir elevio.MotorDirection){
	//Sletter alle ordre med oppgitt etasje i.
	//Kan evt bare slette dem med gitt retning, men er det vits?
	for i := 0; i< _queueSize-2; i++{
		if OrderQueue[i].Floor == floor { //&& (OrderQueue[i].Dir == btn|| OrderQueue[i].Dir == elevio.BT_Cab) {
			OrderQueue[i] = OrderQueue[i+1]
		}
	}

	OrderQueue[_queueSize-1].Floor = -1
	OrderQueue[_queueSize-1].Dir = 0

	fmt.Printf("Order removed. \n New order queue: %+v\n", OrderQueue)
}


func CheckStop(floor int, dir elevio.MotorDirection) bool{	
	var btn elevio.ButtonType
	if dir == elevio.MD_Up{
		btn = elevio.BT_HallUp
	} else if dir == elevio.MD_Stop{
		btn = elevio.BT_Cab
	} else if dir == elevio.MD_Down{
		btn = elevio.BT_HallDown
	}
	return OrderQueue[0].Floor == floor && (OrderQueue[0].Dir == btn || OrderQueue[0].Dir == elevio.BT_Cab) 
}
