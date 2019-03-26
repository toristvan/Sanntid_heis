package queue

import (
	"../driverModule/elevio"
	"../networkModule/bcast"
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

func InitQueue(){
	var invalidOrder config.OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	que := make([][]config.OrderStruct, config.Num_elevs)
	for n := range que{
		que[n] = make([]config.OrderStruct, (elevio.Num_floors*3)-2)
	}
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			que[j][i] = invalidOrder
		}
	}
	orderQueue = que
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
  return 4 - config.LocalID
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
func addToQueue(order config.OrderStruct, current_state config.ElevStateType , id int) { 

	if orderQueue[id][0].Floor == -1{
		insertToQueue(order, 0, id)
	} else if current_state == config.GoingUp {
		if order.Floor < orderQueue[id][0].Floor {
			insertToQueue(order, 0, id)
		}
	} else if current_state == config.GoingDown {
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

	fmt.Printf("Order added\n")
	if !(order.Button == config.BT_Cab && id != config.LocalID){
		elevio.SetButtonLamp(order.Button, order.Floor, true)
	}
}

//Sletter alle ordre med oppgitt etasje i.
//Kan evt bare slette dem med gitt retning, men er det vits?
func RemoveOrder(floor int, id int){
	var prev config.OrderStruct
	//Remove for all id's? Have to beware of cab calls.
	for i := 0; i < queue_size; i++{
		if  orderQueue[id][i].Floor == floor {   //Remove all orders on floor for ID
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

	fmt.Printf("Order removed from queue\n")
	/*for i := 0; i < 10; i++ {
		fmt.Println(orderQueue[1][i].Floor)
		fmt.Println(orderQueue[1][i].Button)
	}*/
	
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

func DistributeOrder(distr_order_chan <-chan config.OrderStruct, add_order_chan chan<- config.OrderStruct, delete_order_chan <-chan config.OrderStruct /*, trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct,*/){
	var lowest_cost int = 10 //max cost
	var best_elev int 	=-1
	var port int 		= 20007
	var new_order config.OrderStruct

	trans_order_chan	:= make (chan config.OrderStruct)
	rec_order_chan		:= make (chan config.OrderStruct)
	trans_conf_chan		:= make (chan config.OrderStruct)
	offline_alert 		:= make (chan bool)

	go bcast.Receiver(port,rec_order_chan)
	go bcast.Transmitter(port, offline_alert, trans_order_chan)
	go bcast.Transmitter(port, offline_alert, trans_conf_chan)  //La inn egen channel for å sende fra Receiverenden, slik at de ikke krasjer. Dårlig løsning?
	//Seems to be many unneccessary if's here
	for{
		ticker := time.NewTicker(100*time.Millisecond) //Need to change this logic
		defer ticker.Stop()
		select{
		case new_order = <- distr_order_chan:
			fmt.Println("distr_order_chan fetched")
			switch new_order.Cmd{
			case config.CostReq:
				new_order.Cmd = config.CostSend
				master = true
				//trans_main_chan <- new_order
				trans_order_chan <- new_order
				fmt.Println("trans_order_chan written to 1")
			case config.OrdrAdd:
				add_order_chan <- new_order
				new_order.Cmd = config.OrdrConf
				new_order.Cost = -1 //cabcall
				//trans_main_chan <- new_order
				trans_order_chan <- new_order
				fmt.Println("trans_order_chan written to 2")

				/*
			case config.OrdrDelete:
				//Will have LocalID
				fmt.Println("Deleting order")
				trans_order_chan <- new_order
				fmt.Println("Order sent for deletion")
				*/
			}
		case new_order = <- delete_order_chan:
			if new_order.Cmd == config.OrdrDelete{
				fmt.Println("Deleting order")
				trans_order_chan <- new_order
				fmt.Println("Order sent for deletion")				
			} else {
				fmt.Println("Wrong command")
			}

		case new_order = <-rec_order_chan:
			fmt.Println("rec_order_chan fetched")

			switch new_order.Cmd{
			case config.CostSend:
				new_order.Cost = dummyCostFunc(new_order)//Current CF requires currentfloor and direction. How to fix
				new_order.ElevID = config.LocalID
				new_order.Cmd = config.OrdrAssign
				//trans_main_chan <- new_order
				trans_conf_chan <- new_order //transmit order cost
				fmt.Println("trans_conf_chan written to")

			case config.OrdrAssign:
				if new_order.MasterID == config.LocalID && new_order.Cost < lowest_cost{
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
					fmt.Println("best_elev", best_elev)
				}
			case config.OrdrAdd:
				if new_order.ElevID == config.LocalID{
					add_order_chan <- new_order //add order to queue
					new_order.Cmd = config.OrdrConf
					//trans_main_chan <- new_order
					trans_conf_chan <- new_order //transmit order confirmation
				}
			case config.OrdrConf:
				if new_order.ElevID != config.LocalID{
					add_order_chan <- new_order
				}
			case config.OrdrDelete:
				if new_order.ElevID != config.LocalID {
					RemoveOrder(new_order.Floor, new_order.ElevID)
				}
			}
		case <- offline_alert:           //To retrieve any offlinemessages blocking. Find better solution
		case <- ticker.C:
			if master {
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_conf_chan <- new_order //transmit new order
				master = false
				lowest_cost = 10 //maxcost
				best_elev = -1
			}
		}
	}
}



func SlaveDistribution(add_order_chan chan<- config.OrderStruct){

	var lowest_cost int = 10 //max cost
	var best_elev int 	=-1
	var master bool 	= false
	var port int 		= 20007

	rec_order_chan		:= make (chan config.OrderStruct)
	trans_conf_chan		:= make (chan config.OrderStruct)
	offline_alert 		:= make (chan bool)

	go bcast.Receiver(port,rec_order_chan)
	go bcast.Transmitter(port, offline_alert, trans_conf_chan)  //La inn egen channel for å sende fra Receiverenden, slik at de ikke krasjer. Dårlig løsning?
	for {
		ticker := time.NewTicker(100*time.Millisecond) //Need to change this logic
		select{
			case new_order = <-rec_order_chan:
			fmt.Println("rec_order_chan fetched")

			switch new_order.Cmd{
				case config.CostSend:
					new_order.Cost = dummyCostFunc(new_order)//Current CF requires currentfloor and direction. How to fix
					new_order.ElevID = config.LocalID
					new_order.Cmd = config.OrdrAssign
					trans_conf_chan <- new_order //transmit order cost
					fmt.Println("trans_conf_chan written to")

				case config.OrdrAssign:
				if new_order.MasterID == config.LocalID && new_order.Cost < lowest_cost{
						master = true
						lowest_cost = new_order.Cost
						best_elev = new_order.ElevID
						fmt.Println("best_elev", best_elev)
					}
				case config.OrdrAdd:
					if new_order.ElevID == config.LocalID{
						add_order_chan <- new_order //add order to queue
						new_order.Cmd = config.OrdrConf
						trans_conf_chan <- new_order //transmit order confirmation
					}
				case config.OrdrConf:
					if new_order.ElevID != config.LocalID{
						add_order_chan <- new_order
					}
				case config.OrdrDelete:
					if new_order.ElevID != config.LocalID {
						RemoveOrder(new_order.Floor, new_order.ElevID)
					}

			case <- offline_alert:           //To retrieve any offlinemessages blocking. Find better solution
			
			case <- ticker.C:
			if master { //Replace with if lowest_cost<10?
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_conf_chan <- new_order //transmit new order
				master = false
				lowest_cost = 10 //maxcost
				best_elev = -1
			}
			}
		}
	}


}
//In channels: drv_buttons (add order) , floor reached (remove order) , costfunction. Out : push Order
func Queue(raw_order_chan <-chan config.OrderStruct, distr_order_chan chan<- config.OrderStruct, add_order_chan <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct/*, trans_main_chan chan<- config.OrderStruct, rec_main_chan <-chan config.OrderStruct*/) {

	//InitQueue()
	//var prev_local_order config.OrderStruct

	//add_to_queue := make(chan config.OrderStruct)
	//distr_order := make(chan config.OrderStruct)
	//source_trans_chan := make(chan config.OrderStruct, 10)
	//sink_rec_chan   := make(chan config.OrderStruct, 10)	
	
	//watchdog_chan := make(chan config.OrderStruct)

	//go watchdog.Watchdog(watchdog_chan, num_elevs, queue_size)
	//maybe necessary to move to main

	//go DistributeOrder(distr_order, add_to_queue, source_trans_chan, sink_rec_chan, config.LocalID)
	//go bcast.OrderReceiver()

	//drv_buttons := make (chan config.ButtonEvent) //distrubieres fra IO_channels, evt kan man kalle IO.IOwrapper er eller noe
	//defer close(drv_buttons)
	//go elevio.PollButtons(drv_buttons)

	for {
		select{
		case new_order := <- raw_order_chan:

			fmt.Printf("Button input: %+v , Floor: %+v\n", new_order.Button, new_order.Floor)
			if !checkIfInQueue(new_order){
				distr_order_chan <- new_order
				fmt.Printf("Distr_order_chan written to\n")
			}

		case order_to_add := <-add_order_chan:
			addToQueue(order_to_add, fsm.RetrieveElevState(), order_to_add.ElevID) //Set lights
			if order_to_add.ElevID == config.LocalID{
				execute_chan <- order_to_add
			}
/*	
		case tmp := <- rec_main_chan:
			fmt.Println("rx queue",tmp)
			sink_rec_chan  <- tmp 

		case tmp := <- source_trans_chan:
			fmt.Println("tx queue",tmp)
			trans_main_chan <- tmp
		*/
		}
	}
}

