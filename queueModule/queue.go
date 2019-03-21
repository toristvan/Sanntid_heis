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
type OrderStruct struct
{
	Button elevio.ButtonType
	Floor int
	timestamp time.Time
}

// flytte til main?
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

func addToQueue(order OrderStruct, id int) {
	fmt.Printf("%+v\n", order)
	
	for i := 0; i< queue_size; i++{
		if orderQueue[id][i].Floor == -1 {
			orderQueue[id][i] = order
			break
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
	var prev orderStruct

	orderQueue[id][0].Floor = -1
	orderQueue[id][0].Button = 0

	for i := 1; i < queue_size ; i++ {
		orderQueue[id][i] = prev
	}

	orderQueue[id][queue_size-1].Floor = -1
	orderQueue[id][queue_size-1].Button = 0

	fmt.Printf("Order removed. \n New order queue: %+v\n", orderQueue)
}

/*
func CheckStop(floor int, dir elevio.MotorDirection, id int) bool{	
	var btn elevio.ButtonType
	if dir == elevio.MD_Up{
		btn = elevio.BT_HallUp
	} else if dir == elevio.MD_Stop{
		btn = elevio.BT_Cab
	} else if dir == elevio.MD_Down{
		btn = elevio.BT_HallDown
	}
	return orderQueue[id][0].floor == floor && (orderQueue[id][0].Button == btn || orderQueue[id][0].Button == elevio.BT_Cab) 
}
*/

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
			
			//Add to watchdog?
			//if new_order.button
			//broadcast_costrequest <- new_order
			//else cabcall
			//broadcast_cabcall
			//addToQueue(new_order, localID)

			//Unnecessary below?
			//assignedorder <- order_assigned
			addToQueue(new_order, localID)

        //case orderreceived
		default:
			if orderQueue[localID][0] != prev_local_order && orderQueue[localID][0].Floor != -1 {
				prev_local_order = orderQueue[localID][0]
				order_chan <- orderQueue[localID][0]
			}
		}
	}
}