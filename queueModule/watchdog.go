package queue

import(
    "../configPackage"
    ."fmt"
    "time"
)

const timeout_threshold time.Duration = 20*time.Second

//Checks if time since order added to queue exceeds timeout_threshold
func timeout (order config.OrderStruct) bool{
    return (time.Since(order.Timestamp)>timeout_threshold);
}

//Iterates through queue every 5 sec, and retransmits timed out orders belonging to other elevators
func Watchdog(distr_order_chan chan<- config.OrderStruct){
    var order_to_retransmit config.OrderStruct  
    for{
        <-time.After(5*time.Second)
        for i := 0; i< config.Num_elevs; i++{
            if i != config.Local_ID {
                for j := 0; j< config.Queue_size; j++{
                    if (orderQueue[i][j].Floor> -1) && (orderQueue[i][j].Floor < config.Num_floors) &&  timeout(orderQueue[i][j]){
                        Println("Watchdog caught a timeout!")
                        order_to_retransmit = orderQueue[i][j]
                        switch orderQueue[i][j].Button {
                        case config.BT_Cab:
                            Printf("Remember to complete your cabcall in floor %d, Elev %d :)\n", order_to_retransmit.ElevID, i)
                            order_to_retransmit.Cmd = config.OrdrRetrans
                        	distr_order_chan <- order_to_retransmit
                        case config.BT_HallUp, config.BT_HallDown:
                            orderQueue[i][j] = invalidateOrder(orderQueue[i][j])
                            Println("'Twas a hall call. I shall do it myself!")
	                        order_to_retransmit.Cmd = config.OrdrRetrans
	                        distr_order_chan <- order_to_retransmit
                        }
                    }
                }
            }
        }
    }
}
