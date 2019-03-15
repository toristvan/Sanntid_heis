package queue

import (
	"./../driverModule/elevio"
	"fmt"
	"time"
	)

const num_elevs int = 1
const queue_size int = (elevio.Num_floors*3)-2
// type BT_Cab signalizises cab call
// maybe add floorstop array
type OrderStruct struct
{
	btn elevio.ButtonType
	floor int
	timestamp time.Time
}

// flytte til main?
var orderQueue [num_elevs][queue_size] OrderStruct
func InitQueue(){
	var invalidOrder OrderStruct
	invalidOrder.btn = 0
	invalidOrder.floor = -1
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			orderQueue[j][i] = invalidOrder
		}
	}
	orderQueue[0][0].btn = 1
	orderQueue[0][0].floor = 0
	orderQueue[0][0].timestamp = time.Now()
}
//var orderQueue := make([] ,queue_size )

func costFunction(newFloor int, currentFloor int, dir elevio.MotorDirection ) int { // IN: currentFloor, Direction, (Queues?) , OUT: Cost
	floorDiff := (newFloor - currentFloor)
	cost := floorDiff
	if floorDiff*int(dir) > 0 && dir != 0 {
		cost = floorDiff - 1
	} else if floorDiff*int(dir) < 0 && dir != 0 {
		cost = floorDiff + 1
	} else {
		cost = floorDiff
	}
	//Broadcast result with ID
	return cost
}

func AddOrder(order queue.OrderStruct, id int) {
	fmt.Printf("%+v\n", order)
	
	}
	for i := 0; i< queue_size; i++{
		if orderQueue[id][i].floor == -1 {
			orderQueue[id][i] = order
			break
		}
	}
	elevio.SetButtonLamp(new_order.Button, new_order.Floor, true)	//Set lamp after order is sent to watchdog
}

func CreateOrder(floor int, btn elevio.ButtonType) OrderStruct{
	var order OrderStruct
	order.btn = btn
	order.floor = floor

	fmt.Printf("Order added: %+v\n Order queue: %+v\n", order, orderQueue)
	return order
}

func RemoveOrder(floor int, id int){
	//Sletter alle ordre med oppgitt etasje i.
	//Kan evt bare slette dem med gitt retning, men er det vits?
	for i := 0; i< queue_size-2; i++{
		if orderQueue[id][i].floor == floor { //&& (orderQueue[i].Btn == btn|| orderQueue[i].Btn == elevio.BT_Cab) {
			orderQueue[id][i] = orderQueue[id][i+1]
		}
	}

	orderQueue[id][queue_size-1].floor = -1
	orderQueue[id][queue_size-1].btn = 0

	fmt.Printf("Order removed. \n New order queue: %+v\n", orderQueue)
}


func CheckStop(floor int, dir elevio.MotorDirection, id int) bool{	
	var btn elevio.ButtonType
	if dir == elevio.MD_Up{
		btn = elevio.BT_HallUp
	} else if dir == elevio.MD_Stop{
		btn = elevio.BT_Cab
	} else if dir == elevio.MD_Down{
		btn = elevio.BT_HallDown
	}
	return orderQueue[id][0].floor == floor && (orderQueue[id][0].btn == btn || orderQueue[id][0].btn == elevio.BT_Cab) 
}
