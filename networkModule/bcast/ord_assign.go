package bcast

import{
	"time"
}

/*
type OrderCommand int

const {
	CostReq		OrderStage = 0
	CostSend	OrderStage = 1
	OrdrAssign	OrderStage = 2
	OrdrAdd		OrderStage = 3
	OrdrConf	OrderStage = 4
}
*/

//in main: orderstage_chan := make (chan bcast.OrderStage)


//transmit_order := make(chan queue.OrderStruct)
//drv_buttons := make(chan elevio.ButtonEvent)

//need two goroutines or seperate functions 
//to make sure only channel on button receiving program is altered

func MasterOrder(tra_order_chan <-chan* queue.OrderStruct){
	var currcost int = 10 //maxORder
	var first bool = true
	var best_elev int =-1
	var master bool = false
	for{
		select{
		case new_order := <-tra_order_chan:
			select{
			case new_order.Cmd == queue.CostReq:
				new_order.cdm = queue.CostSend
				master = true
				Transmit(new_order)
			case new_order.Cmd == queue.OrdrAssign:
				if master{
					if first{
						ticker := time.Ticker(100*Millisecond)
						first = false
					}
					if new_order.Cost<currcost{
						currcost = new_order.Cost
						best_elev = new_order.ElevID
					}
					if ticker.c{
						new_order.ElevID = best_elev
						new_order.Cmd = queue.OrdrAdd
						new_order.Cost = currcost
						Transmit(new_order)
						master = false
						first = true
						currcost = 10 //maxcost
					}
				}

			}

		}
	}
}


//receive_order := make(chan queue.OrderStruct)
//go bcast.Receiver(16569, receive_order)
//go bcast.ReceiveOrders(receive_order)

func SlaveOrder(rec_order_chan <-chan* queue.OrderStruct, add_order_chan chan*<- queue.OrderStruct){
	var currcost int = 10 //maxcost
	for{
		select{
		case new_order := <-rec_order_chan:
			select{
			case new_order.Cmd == queue.CostSend:
				var cost int = queue.CostFunction(new_order) //type for cost instead of int
				new_order.ElevID = LocalID
				new_order.Cmd = queue.OrdrAssign
				Transmit(new_order) 
			case new_order.Cmd == queue.OrdrAdd && new_order.ElevID == LocalID:
				add_order_chan<- new_order
				new_order.Cmd = queue.OrdrConf
				Transmit(new_order)
			case new_order.Cmd == queue.OrdrConf && new_order.ElevID != LocalID:
				add_order_chan<-new_order
			}
			
		}
	}
}


