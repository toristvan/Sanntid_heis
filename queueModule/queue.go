package queue

import (
	"./../driverModule/elevio"
	"./../fsmModule"
	"./../configPackage"
	"fmt"
	)

const num_elevs int = 1
const queue_size int = (elevio.Num_floors*3)-2
const localID int = 0

var orderQueue [num_elevs][queue_size] config.OrderStruct

func InitQueue(){
	var invalidOrder config.OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			orderQueue[j][i] = invalidOrder
		}
	}
	orderQueue[localID][0].Floor = 0
	orderQueue[localID][0].Button = config.BT_Cab
}
//var orderQueue := make([] ,queue_size )

func costFunction(newFloor int, currentFloor int, dir config.MotorDirection ) int { // IN: currentFloor, Direction, (Queues?) , OUT: Cost
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

func dummyCostFunc(hallCall config.ButtonType, floor int, dir config.MotorDirection) int {
  return 1
}

func insertToQueue(order config.OrderStruct, index int, id int){
	for i := queue_size - 1; i > index; i--{
		orderQueue[id][i] = orderQueue[id][i-1]
	}
	orderQueue[id][index] = order
}

func addToQueue(order config.OrderStruct, current_dir config.ElevStateType ,id int) {          //ElevStateType must be elecatorClient package
	fmt.Printf("%+v\n", order)
	if orderQueue[id][0].Floor == -1{
		insertToQueue(order, 0, id)
	} else if current_dir == config.GoingUp {
		if order.Floor < orderQueue[id][0].Floor {
			insertToQueue(order, 0, id)
		}
	} else if current_dir == config.GoingDown {
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

//Sletter alle ordre med oppgitt etasje i.
//Kan evt bare slette dem med gitt retning, men er det vits?
func RemoveOrder(floor int, id int){
	var prev config.OrderStruct

	orderQueue[id][0].Floor = -1
	orderQueue[id][0].Button = 0

	for i := 0; i < queue_size-2 ; i++ {
		prev = orderQueue[id][i]
		orderQueue[id][i] = orderQueue[id][i+1]
		orderQueue[id][i+1] = prev
	}

	fmt.Printf("Order removed. \n New order queue: %+v\n", orderQueue)
}


func checkIfInQueue(order config.OrderStruct) bool{
	for i := 0; i < num_elevs; i++ {
		for j := 0; j < queue_size; j++ {
			if order.Floor == orderQueue[i][j].Floor && order.Button == orderQueue[i][j].Button {
				return true
			}
		}
	}
	return false
}

func DistributeOrder(start_order_chan <-chan config.OrderStruct, add_order_chan chan<- config.OrderStruct, local_id int){
	var lowest_cost int = 10 //maxORder
	var best_elev int =-1
	var master bool = false
	var port int = 20007
	var new_order config.OrderStruct

	trans_order := make (chan config.OrderStruct)
	rec_order := make (chan config.OrderStruct)
	
	go bcast.Receiver(port, rec_order)
	go bcast.Transmitter(port, trans_order)

	//Seems to be many unneccessary if's here
	for{
		ticker := time.NewTicker(100*time.Millisecond)
		defer ticker.Stop()
		select{
		case new_order = <-start_order_chan:
			if new_order.Cmd == config.CostReq {
				new_order.Cmd = config.CostSend
				master = true
				trans_order <- new_order
			}
		case new_order = <-rec_order:
			switch new_order.Cmd{
			case config.CostSend:
				new_order.Cost = dummyCostFunc(new_order.Button, new_order.Floor, config.MD_Down)      //Current CF requires currentfloor and direction. How to fix
				new_order.ElevID = local_id
				new_order.Cmd = config.OrdrAssign
				trans_order <- new_order //transmit new order
			case config.OrdrAssign:
				if master && new_order.Cost < lowest_cost{
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
				}
			case config.OrdrAdd:
				if new_order.ElevID == local_id{
					add_order_chan <- new_order //add order to queue
					new_order.Cmd = config.OrdrConf
					trans_order <- new_order //transmit new order
				}
			case config.OrdrConf:
				if new_order.ElevID != local_id{
					add_order_chan <- new_order

				}
			}
		case <- ticker.C:
			if master {
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_order <- new_order //transmit new order
				master = false
				lowest_cost = 10 //maxcost
				best_elev = -1
			}
		}
	}
}


func Queue(order_chan chan<- config.OrderStruct) {//In channels: drv_buttons (add order) , floor reached (remove order) , costfunction. Out : push Order
	InitQueue()
	var prev_local_order config.OrderStruct

	drv_buttons := make(chan config.ButtonEvent)
	add_to_queue := make(chan config.OrderStruct)
	start_order := make (chan config.OrderStruct)
	defer close(drv_buttons)

	go elevio.PollButtons(drv_buttons)

	//maybe necessary to move to main
	go DistributeOrder(start_order, add_to_queue, localID)

	//go bcast.OrderAssigning(broadcast_costrequest, order_assigned)
	//go bcast.OrderReceiver()

	for {
		select{
		case button_input := <-drv_buttons:   //Button input from elevator
			//Move outside so it's not decalred multiple times?
			var new_order config.OrderStruct

			new_order.Button = button_input.Button
			new_order.Floor = button_input.Floor
			new_order.Cmd = config.CostReq
			start_order <- new_order

			fmt.Printf("Button input: %+v , Floor: %+v\n", new_order.Button, new_order.Floor)
			if !checkIfInQueue(new_order){
				addToQueue(new_order, fsm.RetrieveElevState(), localID)
			}
		case order_to_add := <-add_to_queue:
			addToQueue(order_to_add, fsm.RetrieveElevState(), order_to_add.ElevID)
			//Add to watchdog?
			//if new_order.button
			//broadcast_costrequest <- new_order
			//else cabcall
			//broadcast_cabcall
			//addToQueue(new_order, localID)
			//If queue added, set turnonlightchannels

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

/*
func Queue(input_queue <-chan OrderStruct, order_chan chan<- OrderStruct) {

	select{
	case input := <-input_queue:
		queueIndex += 1
		if queueIndex == queue_size{
			queueIndex = 1
		}
		orderQueue[input.ElevID][queueIndex].Button 		= input.Button
		orderQueue[input.ElevID][queueIndex].Floor 			= input.Floor
		orderQueue[input.ElevID][queueIndex].Timestamp 	= input.Timestamp

		order_chan <- orderQueue[input.ElevID][queueIndex]
	}
}
*/
