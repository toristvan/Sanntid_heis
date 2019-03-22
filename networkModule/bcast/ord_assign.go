package bcast
/* Move to queue?
import(
	"./../../queueModule"
	"time"
)


func DistributeOrder(start_order_chan <-chan* queue.OrderStruct, add_order_chan chan*<- queue.OrderStruct, local_id int){
	var lowest_cost int = 10 //maxORder
	var first_cost bool = true
	var best_elev int =-1
	var master bool = false
	var port int = 20007

	trans_order := make (chan OrderStruct)
	rec_order := make (chan OrderStruct)
	
	go bcast.Receiver(port, rec_order)
	go bcast.Transmitter(port, trans_order)

	for{
		select{
		case new_order := <-start_order_chan:
			select{
			case new_order.Cmd == queue.CostReq:
				new_order.cdm = queue.CostSend
				master = true
				trans_order <- new_order
			}
		case new_order := <-rec_order:
			switch new_order.Cmd{
			case queue.CostSend:
				var cost int = queue.CostFunction(new_order)
				new_order.ElevID = local_id
				new_order.Cmd = queue.OrdrAssign
				trans_order <- new_order //transmit new order
			case queue.OrdrAssign:
				if master{
					if first_cost{
						ticker := time.Ticker(100*Millisecond)
						first_cost = false
					}
					if new_order.Cost < lowest_cost{
						lowest_cost = new_order.Cost
						best_elev = new_order.ElevID
					}
					if ticker.c{
						new_order.ElevID = best_elev
						new_order.Cost = lowest_cost
						new_order.Cmd = queue.OrdrAdd
						trans_order <- new_order //transmit new order
						master = false
						first_cost = true
						lowest_cost = 10 //maxcost
						best_elev = -1
					}
				}
			case queue.OrdrAdd:
				if new_order.ElevID == local_id{
					add_order_chan <- new_order //add order to queue
					new_order.Cmd = queue.OrdrConf
					trans_order <- new_order //transmit new order
				}
			case queue.OrdrConf:
				if new_order.ElevID != local_id{
					add_order_chan <- new_order

				}
			}
		}
	}
}

*/


