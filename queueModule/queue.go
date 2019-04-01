package queue

import (
	"../driverModule/elevio"
    "../networkModule/peers"
	"../configPackage"
	"fmt"
	"time"
	)

func DistributeOrder(distr_order_chan <-chan config.OrderStruct, execute_chan chan<- config.OrderStruct, delete_order_chan <-chan config.OrderStruct, offline_chan <-chan bool, retransmit_last_order_chan <-chan bool, trans_order_chan chan<- config.OrderStruct){
	var new_order config.OrderStruct
	var offline bool = false
	for{
		select{
		case new_order = <- distr_order_chan:
			if !inQueue(new_order){ 
				switch new_order.Cmd{
				//Ask other elevators for cost value
				case config.CostReq:
					new_order.Cmd = config.CostSend
					if offline{ 
						new_order.ElevID = config.Local_ID
						addToQueue(new_order, true)
						execute_chan <- new_order
					} else{
						trans_order_chan <- new_order
					}
				//Case for cab call. 
				case config.OrdrAdd:
					if offline{
						addToQueue(new_order, true)
						execute_chan <- new_order
					}
					new_order.Cost = -1
					trans_order_chan <- new_order
				//Watchdog retransmission: 
				case config.OrdrRetrans:
					new_order.Cmd = config.OrdrAdd
					//Handle hallcall self
					if new_order.Button != config.BT_Cab {
						new_order.ElevID = config.Local_ID
						execute_chan <- new_order
					} 
					addToQueue(new_order, true)
					trans_order_chan <- new_order
				}
			//Retransmit watchdog cabcall to designated elevator
			} else if new_order.Cmd == config.OrdrRetrans {
				new_order.Cmd = config.OrdrAdd
				trans_order_chan <- new_order
			}
		case new_order = <- delete_order_chan:
			RemoveOrder(new_order.Floor, config.Local_ID)
			trans_order_chan <- new_order
		case <- retransmit_last_order_chan:
			new_order.Cmd = config.CostReq 
			trans_order_chan <- new_order 
		case offline = <- offline_chan: 
			fmt.Println("Offline:", offline)
		}
	}
}



func ReceiveOrder(execute_chan chan<- config.OrderStruct, is_dead_chan <-chan bool, retransmit_last_order_chan chan<- bool, rec_order_chan <-chan config.OrderStruct, trans_conf_chan chan<- config.OrderStruct){
	var lowest_cost int = config.Max_cost
	var best_elev int 	=-1
	var master bool 	= false //For elevator on which button was pressed
	var elev_dead bool  = false
	var new_order config.OrderStruct
	for {
		select{
		case new_order = <-rec_order_chan:
			switch new_order.Cmd{
			case config.CostSend:
				if !elev_dead { 
					new_order.Cost = GenericCostFunction(new_order)
					new_order.ElevID = config.Local_ID
					new_order.Cmd = config.OrdrAssign
					trans_conf_chan <- new_order 
				}
			case config.OrdrAssign:
				//If this elev is the one to assign order 
				if new_order.MasterID == config.Local_ID && new_order.Cost < lowest_cost{  
					master = true
					lowest_cost = new_order.Cost
					best_elev = new_order.ElevID
					fmt.Println("Most optimal elevator: ", best_elev)
				}
			case config.OrdrAdd:
				addToQueue(new_order, false)
				new_order.SenderID = config.Local_ID
				new_order.Cmd = config.OrdrConf
				trans_conf_chan <- new_order
			case config.OrdrConf:
				//Don't turn on lights until at least one (other) elevator have order in queue
				//to be sure of retransmission even though elev with order crashes
				//Make sure that elevator is not yourself (unless you are only active peer)
				if new_order.SenderID != config.Local_ID || len(peers.ActivePeers.Peers) < 2 {
					if !inQueue(new_order){
						addToQueue(new_order, true)
					} else if !(new_order.Button == config.BT_Cab && new_order.ElevID != config.Local_ID){
						elevio.SetButtonLamp(new_order.Button, new_order.Floor, true)	
					}
					if new_order.ElevID == config.Local_ID{ 
						execute_chan <- new_order 
					}
				}
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
		case <- time.After(100*time.Millisecond):
			if master {
				new_order.ElevID = best_elev
				new_order.Cost = lowest_cost
				new_order.Cmd = config.OrdrAdd
				trans_conf_chan <- new_order 
				master = false
				lowest_cost = config.Max_cost
				best_elev = -1
			}
		}
	}
}

