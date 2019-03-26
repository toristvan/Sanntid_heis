package queue

import (
	"../driverModule/elevio"
	//"../networkModule/bcast"
	"../fsmModule"
	"../configPackage"
	"fmt"
	"time"
	)

//const num_elevs int  = config.Num_elevs
//const queue_size int = (elevio.Num_floors*3)-2

var num_elevs int 
var queue_size int 
var orderQueue [][] config.OrderStruct
//var orderQueue [config.Num_elevs][queue_size] config.OrderStruct

func InitDataQueue(){
	num_elevs  = config.Num_elevs
	queue_size = (elevio.Num_floors*3)-2
	orderQueue = InitQueue()
	fmt.Println("Queue Ready")
}

func InitQueue() [][]config.OrderStruct {
	var invalidOrder config.OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	que := make([][]config.OrderStruct, num_elevs)
	for n := range que{
		que[n] = make([]config.OrderStruct, queue_size)
	}
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			que[j][i] = invalidOrder
		}
	}
	return que
}

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

//func dummyCostFunc(hallCall config.ButtonType, floor int, dir config.MotorDirection) int {
func dummyCostFunc(order config.OrderStruct) int {
  return 1
}

func RetriveQueue() [][]config.OrderStruct{
	return orderQueue
}

func insertToQueue(order config.OrderStruct, index int, id int){
	for i := queue_size - 1; i > index; i--{
		orderQueue[id][i] = orderQueue[id][i-1]
	}
	order.Timestamp = time.Now()
	orderQueue[id][index] = order
}

//Replace with motordir?
// Make sure lights are only set when we know order will be executed
// For example, when added to queue.
func addToQueue(order config.OrderStruct, current_dir config.ElevStateType , id int) { 

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

	fmt.Printf("Order added")
	elevio.SetButtonLamp(order.Button, order.Floor, true)
}


//Sletter alle ordre med oppgitt etasje i.
//Kan evt bare slette dem med gitt retning, men er det vits?
func RemoveOrder(floor int, id int){
	var prev config.OrderStruct
	//Remove for all id's? Have to beware of cab calls.
	for i := 0; i < queue_size; i++{
		if  orderQueue[id][i].Floor == floor {
			orderQueue[id][i].Floor = -1
			orderQueue[id][i].Button = 0
			orderQueue[id][i].ElevID = -1
		}
	}
	for i := 0; i < queue_size-2 ; i++ {
		prev = orderQueue[id][i]
		orderQueue[id][i] = orderQueue[id][i+1]
		orderQueue[id][i+1] = prev
	}
	for i := config.BT_HallUp; i <= config.BT_Cab  ; i++{ //
		elevio.SetButtonLamp(i, floor, false) //BT syntax correct?
	}

	fmt.Printf("Order removed from queue")
	
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


func DistributeOrder(distr_order_chan <-chan config.OrderStruct, add_order_chan chan<- config.OrderStruct, trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct,local_id int){
	var lowest_cost int = 10 //max cost
	var best_elev int 	=-1
	var master bool 	= false
	//var port int 		= 20007
	var new_order config.OrderStruct

	//trans_order   := make (chan config.OrderStruct)
	//rec_order 	  := make (chan config.OrderStruct)
	//offline_alert := make (chan bool)

	//go bcast.Receiver(port,rec_order)
	//go bcast.Transmitter(port,offline_alert, trans_order)

	//Seems to be many unneccessary if's here
	for{
		ticker := time.NewTicker(1000*time.Millisecond) //Need to change this logic
		defer ticker.Stop()
		select{
		case new_order = <- distr_order_chan:
			switch new_order.Cmd{
			case config.CostReq:
				new_order.Cmd = config.CostSend
				master = true
				trans_main_chan <- new_order
				//trans_order <- new_order
			case config.OrdrAdd:
				add_order_chan <- new_order
				new_order.Cmd = config.OrdrConf
				new_order.Cost = -1 //cabcall
				trans_main_chan <- new_order
				//trans_order <- new_order
			}
		case new_order = <-rec_main_chan:
			switch new_order.Cmd{
			case config.CostSend:
				new_order.Cost = dummyCostFunc(new_order)//Current CF requires currentfloor and direction. How to fix
				new_order.ElevID = local_id
				new_order.Cmd = config.OrdrAssign
				trans_main_chan <- new_order
				//trans_order <- new_order //transmit order cost
			case config.OrdrAssign:
				if master && new_order.Cost < lowest_cost{
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
					fmt.Println("best_elev", best_elev)
				}
			case config.OrdrAdd:
				if new_order.ElevID == local_id{
					add_order_chan <- new_order //add order to queue
					new_order.Cmd = config.OrdrConf
					trans_main_chan <- new_order
					//trans_order <- new_order //transmit order confirmation
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
				trans_main_chan <- new_order //transmit new order
				master = false
				lowest_cost = 10 //maxcost
				best_elev = -1
			}
		}
	}
}

//In channels: drv_buttons (add order) , floor reached (remove order) , costfunction. Out : push Order
func Queue(input_queue <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct, trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct) {

	//InitQueue()
	//var prev_local_order config.OrderStruct

	add_to_queue := make(chan config.OrderStruct)
	distr_order := make(chan config.OrderStruct)

	source_trans_chan := make(chan config.OrderStruct, 10)
	sink_rec_chan   := make(chan config.OrderStruct, 10)


	//watchdog_chan := make(chan config.OrderStruct)

	//go watchdog.Watchdog(watchdog_chan, num_elevs, queue_size)
	//maybe necessary to move to main

	go DistributeOrder(distr_order, add_to_queue, source_trans_chan, sink_rec_chan, config.LocalID)

	//go bcast.OrderReceiver()

	//drv_buttons := make (chan config.ButtonEvent) //distrubieres fra IO_channels, evt kan man kalle IO.IOwrapper er eller noe
	//defer close(drv_buttons)
	//go elevio.PollButtons(drv_buttons)

	for {
		select{
		case new_order := <- input_queue:

			fmt.Printf("Button input: %+v , Floor: %+v\n", new_order.Button, new_order.Floor)
			if !checkIfInQueue(new_order){
				distr_order <- new_order
			}

		case order_to_add := <-add_to_queue:
			addToQueue(order_to_add, fsm.RetrieveElevState(), order_to_add.ElevID) //Set lights
			if order_to_add.ElevID == config.LocalID{
				execute_chan <- order_to_add
			}

		case tmp := <- rec_main_chan:
			fmt.Println("rx queue",tmp)
			sink_rec_chan  <- tmp 

		case tmp := <- source_trans_chan:
			fmt.Println("tx queue",tmp)
			trans_main_chan <- tmp
		
		}
	}
}

