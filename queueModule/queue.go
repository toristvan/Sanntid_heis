package queue

import (
	"./../driverModule/elevio"
	"fmt"
	"time"
	)

const num_elevs int = 1
const queue_size int = (elevio.Num_floors*3)-2
const localID int = 0

// type BT_Cab signalizises cab call
// maybe add floorstop array
type OrderCommand int 

const (
	CostReq 	OrderCommand = 0
	CostSend 	OrderCommand = 1
	OrdrAssign 	OrderCommand = 2
	OrdrAdd 	OrderCommand = 3
	OrdrConf 	OrderCommand = 4
)

type OrderStruct struct
{
	Button elevio.ButtonType
	Floor int
	timestamp time.Time
	ElevID int
	Cost int
	Cmd OrderCommand

}


var orderQueue [num_elevs][queue_size] OrderStruct


func InitQueue(){
	var invalidOrder OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			orderQueue[j][i] = invalidOrder
		}
	}
	orderQueue[localID][0].Floor = 0
	orderQueue[localID][0].Button = elevio.BT_Cab
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

func insertToQueue(order OrderStruct, index int, id int){
	for i := queue_size - 1; i > index; i--{
		orderQueue[id][i] = orderQueue[id][i-1]
	}
	orderQueue[id][index] = order
}

func addToQueue(order OrderStruct, current_dir ElevStateType ,id int) {          //ElevStateType must be elecatorClient package
	fmt.Printf("%+v\n", order)
	if orderQueue[id][0].Floor == -1{
		insertToQueue(order, 0, id)
	} else if currentDir == GoingUp {
		if order.Floor < orderQueue[id][0].Floor {
			insertToQueue(order, 0, id)
		}
	} else if currentDir == GoingDown {
		if order.Floor > orderQueue[id][0].Floor {
			insertToQueue(order, 0, id)
		}
	} else {
		for i := 0; i < queue_size; i++{
			if orderQueue[id][i].Floor == -1 {
				orderQueue[id][i] = order
				break
			}
		}
	}

	fmt.Printf("Order added. Current queue: %+v\n", orderQueue)
}

/*
func CreateOrder(floor int, btn elevio.ButtonType) OrderStruct{
	var order OrderStruct
	order.Button = btn
	order.Floor = floor

	fmt.Printf("Order added: %+v\n Order queue: %+v\n", order, orderQueue)
	return order
}
*/

func RemoveOrder(floor int, id int){
	//Sletter alle ordre med oppgitt etasje i.
	//Kan evt bare slette dem med gitt retning, men er det vits?
	var prev OrderStruct

	orderQueue[id][0].Floor = -1
	orderQueue[id][0].Button = 0

	for i := 0; i < queue_size-2 ; i++ {
		prev = orderQueue[id][i]
		orderQueue[id][i] = orderQueue[id][i+1]
		orderQueue[id][i+1] = prev
	}

	fmt.Printf("Order removed. \n New order queue: %+v\n", orderQueue)
}


func checkIfInQueue(order OrderStruct) bool{
	for i := 0; i < num_elevs; i++ {
		for j := 0; j < queue_size; j++ {
			if order.Floor == orderQueue[i][j].Floor && order.Button == orderQueue[i][j].Button {
				return true
			}
		}
	}
	return false
}


func Queue(order_chan chan<- OrderStruct) {//In channels: drv_buttons (add order) , floor reached (remove order) , costfunction. Out : push Order
	InitQueue()
	var prev_local_order OrderStruct

	drv_buttons := make(chan elevio.ButtonEvent)
	defer close(drv_buttons)
	//broadcast_costrequest := make(chan )

	go elevio.PollButtons(drv_buttons)

	//go bcast.OrderAssigning(broadcast_costrequest, order_assigned)
	//go bcast.OrderReceiver()

	for {
		select{
		case button_input := <- drv_buttons:   //Button input from elevator
			var new_order OrderStruct

			new_order.Button = button_input.Button
			new_order.Floor = button_input.Floor

			fmt.Printf("Button input: %+v , Floor: %+v\n", new_order.Button, new_order.Floor)
			if !checkIfInQueue(){
				addToQueue(new_order, localID)
			}
			//Add to watchdog?
			//if new_order.button
			//broadcast_costrequest <- new_order
			//else cabcall
			//broadcast_cabcall
			//addToQueue(new_order, localID)

			//Unnecessary below?
			//assignedorder <- order_assigned

        //case orderreceived
		default:
			if orderQueue[localID][0] != prev_local_order && orderQueue[localID][0].Floor != -1 {
				prev_local_order = orderQueue[localID][0]
				order_chan <- orderQueue[localID][0]
			}
		}
	}
}