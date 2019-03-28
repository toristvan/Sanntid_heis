package queue

import(
  //"../networkModule/bcast"
  "../configPackage"
  ."fmt"
  "time"
  )

const timeout_threshold time.Duration = 10*time.Second

func timeout (order config.OrderStruct) bool{
  return time.Since(order.Timestamp)>timeout_threshold;
}

func Watchdog(distr_order_chan chan<- config.OrderStruct){
  //a := NewWatchdog()
  //a.initWatchdog(10 * time.Second)
  for{
    time.Sleep(5 * time.Second)
    //a.updateWatchdog()
    //a.resetWatchdog()
    for i := 0; i< config.Num_elevs; i++{
      if i != config.LocalID {
        for j := 0; j< Queue_size; j++{
          if orderQueue[i][j].Floor != -1 && timeout(orderQueue[i][j]){
            Println("Watchdog caught HIV")
            order_to_retransmit := orderQueue[i][j]
            switch orderQueue[i][j].Button {
            case config.BT_HallUp, config.BT_HallDown:
              orderQueue[i][j] = invalidateOrder(orderQueue[i][j])
            case config.BT_Cab:
            }
            order_to_retransmit.Cmd = config.OrdrRetrans
            distr_order_chan <- order_to_retransmit
          }
        }
      }
    }
  }
}
