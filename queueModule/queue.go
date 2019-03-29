package queue

import (
	"../driverModule/elevio"
	//"../networkModule/bcast"
    "../networkModule/peers"
	"../fsmModule"
	"../configPackage"
	"fmt"
	"time"
	)

//Move to config
const Queue_size int = (config.Num_floors*3)-2

//var orderQueue [][] config.OrderStruct
var orderQueue [config.Num_elevs][Queue_size] config.OrderStruct

func InitQueue(){
	for j := 0; j< config.Num_elevs; j++{
		for i := 0; i< Queue_size; i++{
			orderQueue[j][i] = invalidateOrder(orderQueue[j][i]) 
			orderQueue[j][i].ElevID = j
		}
	}
}

func invalidateOrder(order config.OrderStruct) config.OrderStruct {
	order.Button = config.BT_HallUp
	order.Floor = -1
	order.Cost = config.Max_cost
	order.Cmd = config.OrdrInv
	order.MasterID = -1
	order.SenderID = -1
	return order
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
			printOrder(orderQueue[config.Local_ID][i])		
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
  		cost = abs_distance - 3 - config.Num_floors //best option
  	case false:
    	cost =  abs_distance - 1 - config.Num_floors //config.Num_floors + 1 - abs_distance
  	}
  case config.AtFloor:
    cost =  abs_distance - config.Num_floors//config.Num_floors - abs_distance
  case config.GoingUp:
    switch distance < 0{ 
    case true:
      cost =  abs_distance //- abs_distance //1
      //bad way to measure? because long way down
    case false:
      cost = abs_distance - 2 - config.Num_floors//config.Num_floors + 2 - abs_distance
    }
  case config.GoingDown:
    switch distance < 0{
    case true:
      cost = abs_distance - 2 - config.Num_floors//config.Num_floors + 2 - abs_distance
      //Add what way order is going
    case false:
      //bad way to measure? because long way up
      cost = abs_distance //- abs_distance //1
    }
  }

  //Add whether or not there are cab calls?
  //Take into consideration what direction order is going?
  //Works surprisingly well without these factors
  fmt.Printf("Cost for %d: %d\n", config.Local_ID, cost)
  return cost
}

func RetrieveQueue() [config.Num_elevs][Queue_size]config.OrderStruct{
	return orderQueue
}

func insertToQueue(order config.OrderStruct, index int){
	for i := Queue_size - 1; i > index; i--{
		orderQueue[order.ElevID][i] = orderQueue[order.ElevID][i-1]
	}
	order.Timestamp = time.Now()
	//printOrder(order)
	orderQueue[order.ElevID][index] = order
}

//Replace with motordir?
// Make sure lights are only set when we know order will be executed
// For example, when added to queue.
func addToQueue(order config.OrderStruct, set_lights bool) {
	current_state := fsm.RetrieveElevState()
	if orderQueue[order.ElevID][0].Floor == -1{
		insertToQueue(order, 0)
	} else if current_state == config.GoingUp && order.ElevID == config.Local_ID{
		if order.Floor < orderQueue[order.ElevID][0].Floor {
			insertToQueue(order, 0)
		}
	} else if current_state == config.GoingDown && order.ElevID == config.Local_ID{
		if order.Floor > orderQueue[order.ElevID][0].Floor {
			insertToQueue(order, 0)
		}
	} else {
		for i := 0; i < Queue_size; i++{
			if orderQueue[order.ElevID][i].Floor == -1 {
				insertToQueue(order, i)
				break
			}
		}
	}
	//fmt.Printf("Order added\n")
	if set_lights && !(order.Button == config.BT_Cab && order.ElevID != config.Local_ID){
		elevio.SetButtonLamp(order.Button, order.Floor, true)
	}
}

func RemoveOrder(floor int, id int){
	for i := 0; i < config.Num_elevs; i++ {
		for j := 0; j < Queue_size; j++{
			//Remove all olders on floor except cab calls not on ID
			if  orderQueue[i][j].Floor == floor && (orderQueue[i][j].Button != config.BT_Cab || id == i) { 
				orderQueue[i][j] = invalidateOrder(orderQueue[i][j])
			}
		}
	}
	if id == config.Local_ID {
		elevio.SetButtonLamp(config.ButtonType(config.BT_Cab), floor, false)
	} 
	for i := config.BT_HallUp; i < config.BT_Cab  ; i++{ //
		elevio.SetButtonLamp(config.ButtonType(i), floor, false) 
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
			if order.Floor == orderQueue[config.Local_ID][j].Floor  && order.Button == orderQueue[config.Local_ID][j].Button{
				return true
			}
		}			
	}
	return false
}

func DistributeOrder(distr_order_chan <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct, delete_order_chan <-chan config.OrderStruct, offline_chan <-chan bool, retransmit_last_order_chan <-chan bool, trans_order_chan chan<- config.OrderStruct){
	var new_order config.OrderStruct
	var offline bool = false
	//Make order chan bigger, so not freeze as easily?
	for{
		select{
		case new_order = <- distr_order_chan:
			//fmt.Println("DistributeOrder")
			if !checkIfInQueue(new_order){   //Should only check for own queue
				switch new_order.Cmd{
				case config.CostReq:
					new_order.Cmd = config.CostSend
					if offline{ //Take order self if offline
						new_order.ElevID = config.Local_ID //In case of retransmission
						addToQueue(new_order, true)
						execute_chan <- new_order
					} else{
						trans_order_chan <- new_order
					}
				case config.OrdrAdd:
					if offline{
						addToQueue(new_order, true)
						execute_chan <- new_order
					}
					new_order.Cost = -1 //cabcall
					trans_order_chan <- new_order
				//Add watchdog hallcall to own queue, and transmit adding
				case config.OrdrRetrans:
					new_order.Cmd = config.OrdrAdd
					if new_order.Button != config.BT_Cab {
						new_order.ElevID = config.Local_ID
						execute_chan <- new_order
					} //If cab call, will belong to other elevator. Retransmit and add to queue.
					addToQueue(new_order, true)
					trans_order_chan <- new_order
				}
			//Retransmit watchdog cabcall to designated elevator
			} else if new_order.Cmd == config.OrdrRetrans {
				new_order.Cmd = config.OrdrAdd
				trans_order_chan <- new_order
			}
		case new_order = <- delete_order_chan:
			if new_order.Cmd == config.OrdrDelete{
				trans_order_chan <- new_order
			}
		case <- retransmit_last_order_chan:
			new_order.Cmd = config.CostReq //Added newly
			trans_order_chan <- new_order //what if new_order changes? What if cab call?
		case offline = <- offline_chan: //sets offline if transmit cant connect to router
			fmt.Println("Offline:", offline)
		}
	}
}



func ReceiveOrder(execute_chan chan<- config.OrderStruct, is_dead_chan <-chan bool, retransmit_last_order_chan chan<- bool, rec_order_chan <-chan config.OrderStruct, trans_conf_chan chan<- config.OrderStruct){
	var lowest_cost int = config.Max_cost //max cost
	var best_elev int 	=-1
	var master bool 	= false
	var elev_dead bool  = false
	var new_order config.OrderStruct
	for {
		//assign_timeout := time.NewTicker(100*time.Millisecond) //Need to change this logic
		//defer assign_timeout.Stop()
		select{
		case new_order = <-rec_order_chan:
			switch new_order.Cmd{
			//Send cost value for given order
			case config.CostSend:
				if !elev_dead { //don't send cost if dead
					new_order.Cost = GenericCostFunction(new_order)
					new_order.ElevID = config.Local_ID
					new_order.Cmd = config.OrdrAssign
					trans_conf_chan <- new_order //transmit order cost					
				}
			case config.OrdrAssign:
				//If this elev is the one to assign order.
				if new_order.MasterID == config.Local_ID && new_order.Cost < lowest_cost{  
					master = true
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
					//fmt.Println("Most optimal elevator: ", best_elev)
				}
			//Confirm order receival
			case config.OrdrAdd:
				//fmt.Println("Received add request")
				//Checkif in queue?
				addToQueue(new_order, false)
				new_order.SenderID = config.Local_ID
				new_order.Cmd = config.OrdrConf
				trans_conf_chan <- new_order //transmit order confirmation
			//Add to queue and 
			case config.OrdrConf:
				//fmt.Println("Received order confirmation")
				//Don't turn on lights until at least two elevators have order in queue
				//to be sure of retransmission even though elev with order crashes
				if new_order.SenderID != config.Local_ID || len(peers.ActivePeers.Peers) < 2 {
					if !checkIfInQueue(new_order){
						addToQueue(new_order, true)
					} else if !(new_order.Button == config.BT_Cab && new_order.ElevID != config.Local_ID){
						elevio.SetButtonLamp(new_order.Button, new_order.Floor, true)	
					}
					if new_order.ElevID == config.Local_ID{ //if order belongs to this elev
						execute_chan <- new_order //add to stopArray
					}
				}
			//Other elevator has finished its order
			case config.OrdrDelete:
				if new_order.ElevID != config.Local_ID {
					RemoveOrder(new_order.Floor, new_order.ElevID)
				}
			}
		//If 'dead' e.g motor unplugged
		case elev_dead = <- is_dead_chan: 
			fmt.Println("Dead:", elev_dead)
			if elev_dead{
				retransmit_last_order_chan <- true
			}
		//Decide who gets order after certain time
		case <- time.After(100*time.Millisecond)/*assign_timeout.C*/:
			if master {
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_conf_chan <- new_order //transmit new order
				master = false
				lowest_cost = config.Max_cost //maxcost
				best_elev = -1
			}
		}
	}
}

/*
func SpamOrder(spam_chan chan<- config.OrderStruct){
	var dummyorder config.OrderStruct
	dummyorder.ElevID = config.Local_ID
    dummyorder.Button    = config.BT_Cab
    dummyorder.MasterID  = config.Local_ID
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
