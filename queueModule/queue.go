package queue

import (
	"../driverModule/elevio"
	"../networkModule/bcast"
	"../fsmModule"
	"../configPackage"
	"fmt"
	"time"
	)

//Move to config
const Queue_size int = (elevio.Num_floors*3)-2

//var orderQueue [][] config.OrderStruct
var orderQueue [config.Num_elevs][Queue_size] config.OrderStruct

func InitQueue(){
	var invalidOrder config.OrderStruct
	invalidOrder.Button = 0
	invalidOrder.Floor = -1
	invalidOrder.Cmd = 1
	/*
	que := make([][]config.OrderStruct, config.Num_elevs)
	for n := range que{
		que[n] = make([]config.OrderStruct, (elevio.Num_floors*3)-2)
	}*/
	for j := 0; j< config.Num_elevs; j++{
		for i := 0; i< Queue_size; i++{
			orderQueue[j][i] = invalidOrder
			orderQueue[j][i].ElevID = j
		}
	}
}

func printOrder(order config.OrderStruct){
  	fmt.Printf("\nID: %d Button: %d Floor: %d Cost: %d Cmd: %d Time: %s\n", order.ElevID, order.Button, order.Floor, order.Cost, order.Cmd, order.Timestamp.String())
  	//fmt.Println("\nID:",order.ElevID, "  Button:", order.Button, " Floor: ", order.Floor," Cost: ", order.Cost," Cmd: ", order.Cmd)
	//fmt.Printf("%+v", orderQueue)
}
func PrintQueue(){
	for {
		//for j := 0; j < config.Num_elevs ; j++{
		for i := 0 ; i < len(orderQueue[0]) ;  i++{
			printOrder(orderQueue[config.LocalID][i])		
		}
		//} 
		time.Sleep(5*time.Second)
	}
}


//Might have weaknesses considering state changes
func GenericCostFunction(order config.OrderStruct) int {
  var cost int
  var distance int = config.Current_floor - order.Floor
  var abs_distance int
  if distance < 0 {
    abs_distance = -distance
  }else{
    abs_distance = distance
  }
  //in outcommented values; higher cost i better
  //Put in max cost to make sure always pos?
  switch fsm.RetrieveElevState(){
  case config.Idle:
  	switch distance == 0{
  	case true:
  		cost = abs_distance - 3 - elevio.Num_floors //best option
  	case false:
    	cost =  abs_distance - 1 - elevio.Num_floors //config.Num_floors + 1 - abs_distance
  	}
  case config.AtFloor:
    cost =  abs_distance - elevio.Num_floors//config.Num_floors - abs_distance
  case config.GoingUp:
    switch distance < 0{ 
    case true:
      cost =  abs_distance //- abs_distance //1
      //bad way to measure? because long way down
    case false:
      cost = abs_distance - 2 - elevio.Num_floors//config.Num_floors + 2 - abs_distance
    }
  case config.GoingDown:
    switch distance < 0{
    case true:
      cost = abs_distance - 2 - elevio.Num_floors//config.Num_floors + 2 - abs_distance
      //Add what way order is going
    case false:
      //bad way to measure? because long way up
      cost = abs_distance //- abs_distance //1
    }
  }

  //Add whether or not there are cab calls?
  //Take into consideration what direction order is going?
  //Works surprisingly well without these factors
  fmt.Printf("Cost for %d: %d\n", config.LocalID, cost)
  return cost
}

//func dummyCostFunc(hallCall config.ButtonType, floor int, dir config.MotorDirection) int {
func dummyCostFunc(order config.OrderStruct) int {
  return 4 - config.LocalID
}
func SetQueue(new_queue [config.Num_elevs][Queue_size]config.OrderStruct){
	orderQueue = new_queue
}
func RetrieveQueue() [config.Num_elevs][Queue_size]config.OrderStruct{
	return orderQueue
}

func insertToQueue(order config.OrderStruct, index int, id int){
	for i := Queue_size - 1; i > index; i--{
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
		for i := 0; i < Queue_size; i++{
			if orderQueue[id][i].Floor == -1 {
				orderQueue[id][i] = order
				break
			}
		}
	}

	//fmt.Printf("Order added\n", orderQueue)
	if !(order.Button == config.BT_Cab && id != config.LocalID){
		elevio.SetButtonLamp(order.Button, order.Floor, true)
	}
}

//Sletter alle ordre med oppgitt etasje i.
//Kan evt bare slette dem med gitt retning, men er det vits?
func RemoveOrder(floor int, id int){
	var prev config.OrderStruct
	//Remove for all id's? Have to beware of cab calls.
	for i := 0; i < Queue_size; i++{
		if  orderQueue[id][i].Floor == floor {   //Remove all orders on floor for ID
			orderQueue[id][i].Floor = -1
			orderQueue[id][i].Button = 0
			orderQueue[id][i].ElevID = -1
		}
	}
	for i := 0; i < Queue_size-2 ; i++ {
		prev = orderQueue[id][i]
		orderQueue[id][i] = orderQueue[id][i+1]
		orderQueue[id][i+1] = prev
	}
	for i := config.BT_HallUp; i <= config.BT_Cab  ; i++{ //
		elevio.SetButtonLamp(i, floor, false) //BT syntax correct?
	}

	//fmt.Printf("Order removed from queue\n")
	/*for i := 0; i < 10; i++ {
		fmt.Println(orderQueue[1][i].Floor)
		fmt.Println(orderQueue[1][i].Button)
	}*/

}

func checkIfInQueue(order config.OrderStruct) bool{
	if order.Button != config.BT_Cab{
		for i := 0; i < config.Num_elevs; i++ {
			for j := 0; j < Queue_size; j++ {
				if order.Floor == orderQueue[i][j].Floor && order.Button == orderQueue[i][j].Button  {
					return true
				}
			}
		}
	}else{
		for j := 0; j < Queue_size; j++ {
			if order.Floor == orderQueue[config.LocalID][j].Floor  && order.Button == orderQueue[config.LocalID][j].Button{
				return true
			}
		}			
	}
	return false
}

func DistributeOrder(is_dead_chan <-chan bool, distr_order_chan <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct, delete_order_chan <-chan config.OrderStruct, offline_chan <-chan bool){
	var new_order config.OrderStruct
	var offline bool = false
	trans_order_chan	:= make (chan config.OrderStruct)
	go bcast.Transmitter(config.Order_port, trans_order_chan)

	//Make order chan bigger, so not freeze as easily?
	for{
		select{
		case <- is_dead_chan:
			trans_order_chan <- new_order
		case new_order = <- distr_order_chan:
			fmt.Println("DistributeOrder")
			if !checkIfInQueue(new_order){
				switch new_order.Cmd{
				case config.CostReq:
					new_order.Cmd = config.CostSend
					if offline{ //Take order self if offline
						new_order.ElevID = config.LocalID
						//add_order_chan <- new_order
						addToQueue(new_order, fsm.RetrieveElevState(), new_order.ElevID)
						execute_chan <- new_order
					} else{
						trans_order_chan <- new_order
					}
				case config.OrdrAdd:
					//add_order_chan <- new_order
					addToQueue(new_order, fsm.RetrieveElevState(), new_order.ElevID)
					execute_chan <- new_order
					new_order.Cmd = config.OrdrConf
					new_order.Cost = -1 //cabcall
					trans_order_chan <- new_order
				}
			}
		case new_order = <- delete_order_chan:
			if new_order.Cmd == config.OrdrDelete{
				trans_order_chan <- new_order
			}
		//case <- offline_chan:
		case offline = <- offline_chan: //sets offline if transmit cant connect to router
			fmt.Println("offline:", offline)
		}
	}
}



func ReceiveOrder(execute_chan chan<- config.OrderStruct, is_dead_chan <-chan bool){

	var lowest_cost int = config.MaxCost //max cost
	var best_elev int 	=-1
	var master bool 	= false
	var elev_dead bool  = false
	var new_order config.OrderStruct

	rec_order_chan		:= make (chan config.OrderStruct)
	trans_conf_chan		:= make (chan config.OrderStruct)
	trans_backup_chan	:= make (chan bool)

	go bcast.Receiver(config.Order_port, rec_order_chan)
	go bcast.Transmitter(config.Order_port/*, offline_alert_chan*/, trans_conf_chan)  //Channel to send orders
	go bcast.Transmitter(config.Backup_port/*, offline_backup_chan*/, trans_backup_chan)  //Channel to send heartbeat to backup

	for {
		assign_timeout := time.NewTicker(100*time.Millisecond) //Need to change this logic
		defer assign_timeout.Stop()
		select{
		case new_order = <-rec_order_chan:
			switch new_order.Cmd{
			case config.CostSend:
				if elev_dead { //donst send cost if dead
					break
				}
				new_order.Cost = GenericCostFunction(new_order)//dummyCostFunc(new_order)//Current CF requires currentfloor and direction. How to fix
				new_order.ElevID = config.LocalID
				new_order.Cmd = config.OrdrAssign
				trans_conf_chan <- new_order //transmit order cost

			case config.OrdrAssign:
				if new_order.MasterID == config.LocalID && new_order.Cost < lowest_cost{  //If this elev is the one to assign order.
					master = true
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
					fmt.Println("best_elev", best_elev)
				}
			case config.OrdrAdd:
				if new_order.ElevID == config.LocalID{
					addToQueue(new_order, fsm.RetrieveElevState(), new_order.ElevID)
					execute_chan <- new_order //add to stopArray
					//add_order_chan <- new_order //add order to queue
					new_order.Cmd = config.OrdrConf
					trans_conf_chan <- new_order //transmit order confirmation
				}
			case config.OrdrConf:
				if new_order.ElevID != config.LocalID{
					//add_order_chan <- new_order
					addToQueue(new_order, fsm.RetrieveElevState(), new_order.ElevID)
				}
			case config.OrdrDelete:
				if new_order.ElevID != config.LocalID {
					RemoveOrder(new_order.Floor, new_order.ElevID)
				}
			}
		case elev_dead = <- is_dead_chan: //If 'dead' e.g motor unplugged
			fmt.Println("Dead:", elev_dead)

		case <- assign_timeout.C:
			if master { //Replace with if lowest_cost<10?
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_conf_chan <- new_order //transmit new order
				master = false
				lowest_cost = config.MaxCost //maxcost
				best_elev = -1
			}
			trans_backup_chan <- true
		}
	}
}

/*
func SpamOrder(spam_chan chan<- config.OrderStruct){
	var dummyorder config.OrderStruct
	dummyorder.ElevID = config.LocalID
    dummyorder.Button    = config.BT_Cab
    dummyorder.MasterID  = config.LocalID
    dummyorder.Cmd = config.OrdrAdd
	for {
		time.Sleep(5*time.Second)
		dummyorder.Floor = 3
		spam_chan <- dummyorder

		time.Sleep(5*time.Second)
		dummyorder.Floor = 0
		spam_chan <- dummyorder

	}
}
*/
