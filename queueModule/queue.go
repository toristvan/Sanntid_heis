package queue

import (
	"../driverModule/elevio"
	"../networkModule/bcast"
	"../fsmModule"
	"../configPackage"
	"fmt"
	"time"
	)

const num_elevs int = 1
const queue_size int = (elevio.Num_floors*3)-2
//const localID int = 0  Deklarering av denne på flere plasser ser ut til å føre til feil. flyttet til config
//kan eksempelvis få "index out of range"
var orderQueue [config.Num_elevs][queue_size] config.OrderStruct

func InitQueue(){
	var invalidOrder config.OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	for j := 0; j< num_elevs; j++{
		for i := 0; i< queue_size; i++{
			orderQueue[j][i] = invalidOrder
		}
	}
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

//func dummyCostFunc(hallCall config.ButtonType, floor int, dir config.MotorDirection) int {
func dummyCostFunc(order config.OrderStruct) int {
  return 1
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
	//for j := 0; j < queue_size; j++{
		fmt.Printf("Order added")
	//}
	elevio.SetButtonLamp(order.Button, order.Floor, true)

}

/*
func addToQueue(order_to_add <-chan config.OrderStruct, order_added chan<- config.OrderStruct, id int){
	current_dir := fsm.RetrieveElevState()
	select{
	case order := <- order_to_add:
		if !checkIfInQueue(order){
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
			for j := 0; j < queue_size; j++{
				fmt.Printf("Order added. Current queue: %+v\n", orderQueue[id][j])
			}
			order_added <- orderQueue[id][0]
		}
	}
}
*/

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

	//for j := 0; j < queue_size; j++{
		fmt.Printf("Order removed from queue")
	//}
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

func DistributeOrder(distr_order_chan <-chan config.OrderStruct, add_order_chan chan<- config.OrderStruct, local_id int){
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
		case new_order = <- distr_order_chan:
			switch new_order.Cmd{
			case config.CostReq:
				new_order.Cmd = config.CostSend
				master = true
				trans_order <- new_order
			case config.OrdrAdd:
				add_order_chan <- new_order
				new_order.Cmd = config.OrdrConf
				new_order.Cost = -1 //cabcall
				trans_order <- new_order
			}
		case new_order = <-rec_order:
			switch new_order.Cmd{
			case config.CostSend:
				new_order.Cost = dummyCostFunc(new_order)//Current CF requires currentfloor and direction. How to fix
				new_order.ElevID = local_id
				new_order.Cmd = config.OrdrAssign
				trans_order <- new_order //transmit order cost
			case config.OrdrAssign:
				if master && new_order.Cost < lowest_cost{
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
				}
			case config.OrdrAdd:
				if new_order.ElevID == local_id{
					add_order_chan <- new_order //add order to queue
					new_order.Cmd = config.OrdrConf
					trans_order <- new_order //transmit order confirmation
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

func Queue(input_queue <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct) {//In channels: drv_buttons (add order) , floor reached (remove order) , costfunction. Out : push Order
	//InitQueue()
	//var prev_local_order config.OrderStruct

	add_to_queue := make(chan config.OrderStruct)
	distr_order := make(chan config.OrderStruct)
	//watchdog_chan := make(chan config.OrderStruct)

	//go watchdog.Watchdog(watchdog_chan, num_elevs, queue_size)
	//maybe necessary to move to main
	go DistributeOrder(distr_order, add_to_queue, config.LocalID)
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

		default:  //Prøver å lage en logikk som sender ordre til executeOrder (legger til i stopArray) når det er relevant
			//Send ordre til executeOrder hvis du er idle og har ordre i køen
			//Eller hvis du er på vei en vei og øverste ordre skal samme vei
			
			//Per nu sender den ordre hvis det finnes. midlertidig. Må legges inn en algoritme som bestemmer om det skal legges til eller ikke

			//switch fsm.RetrieveElevState() {
			//case config.Idle:
			if orderQueue[config.LocalID][0].Floor != -1 {
				execute_chan <- orderQueue[config.LocalID][0]
			}
			time.Sleep(50*time.Millisecond)


			//Add to watchdog?
			//if new_order.button
			//broadcast_costrequest <- new_order
			//else cabcall
			//broadcast_cabcall
			//addToQueue(new_order, localID)
			//If queue added, set turnonlightchannels

			//Unnecessary below?
			//assignedorder <- order_assigned
			//case re_add_order := <-watchdog_chan:
			//addToQueue(re_add_order, fsm.RetrieveElevState(), localID)
		/*
		case button_input := <-drv_buttons:   //Button input from elevator
			//Move outside so it's not decalred multiple times?
			var new_order config.OrderStruct

			new_order.Button = button_input.Button
			new_order.Floor = button_input.Floor

			// If Hallcall, need to allocate order
			if (new_order.Button != config.BT_Cab){
				new_order.Cmd = config.CostReq
			} else{
				new_order.Cmd = config.OrdrAdd
			}
			start_order <- new_order

			fmt.Printf("Button input: %+v , Floor: %+v\n", new_order.Button, new_order.Floor)
			if !checkIfInQueue(new_order){
				addToQueue(new_order, fsm.RetrieveElevState(), localID)
			}

		case order_to_add := <- input_queue:
			fmt.Println(order_to_add)
			addToQueue(order_to_add, fsm.RetrieveElevState(), order_to_add.ElevID)
			//is this to add cab calls? in that case, make sure to transmit with Cmd=OrdrConf
			//new_order.Cmd = config.OrdrConf
			//start_order <= new__order
			//If not, replacment for whats below?


		default: //Blir noe trøbbel med denne
			if orderQueue[config.LocalID][0] != prev_local_order && orderQueue[config.LocalID][0].Floor != -1 {
				prev_local_order = orderQueue[config.LocalID][0]
				execute_chan <- orderQueue[config.LocalID][0] //ser ut at denne kanskje blokkerer da den ikke er i bruk
			} else {
				time.Sleep(100*time.Millisecond)   //Unload CPU
			}
			*/
		}
	}
}


/*
func Queue(input_queue <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct) {

	for{
		select{
		case order_to_add := <-input_queue:
			fmt.Println("new order")
			addToQueue(order_to_add, fsm.RetrieveElevState(), order_to_add.ElevID)

			queueIndex +=
			if queueIndex == queue_size{
				queueIndex = 1
			}
			orderQueue[input.ElevID][queueIndex].Button 		= input.Button
			orderQueue[input.ElevID][queueIndex].Floor 			= input.Floor
			orderQueue[input.ElevID][queueIndex].Timestamp 	= input.Timestamp

			order_chan <- orderQueue[input.ElevID][queueIndex]

		}
	}
}
*/
