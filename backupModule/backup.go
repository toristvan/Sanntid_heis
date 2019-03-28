package backup

import(
	"../networkModule/bcast"
	"../queueModule"
	"../configPackage"
	//"../driverModule/elevio" //remove when num_floors are removed
	"fmt"
	"time"

)


func RequestBackup(distr_order_chan chan<- config.OrderStruct, backup_req_chan <-chan int, transmit_backup_chan chan<- config.OrderStruct) {
	var backup_queue config.OrderStruct
	var backup_received bool = false
	var buffer [1][queue.Queue_size]config.OrderStruct
	var index_buffer int = 0

	//backup_queue_chan := make(chan [config.Num_elevs][queue.Queue_size]config.OrderStruct)
	backup_queue_chan := make(chan config.OrderStruct)
	request_backup_chan := make(chan int)

	//Transmit backup request and receive backup-queue
	go bcast.Receiver(config.Backup_port, backup_queue_chan)
	go bcast.Transmitter(config.Backup_port, request_backup_chan)
	request_backup_chan <- config.LocalID
	admit_loneliness := time.NewTicker(5*time.Second)

	for{
		select{
		case backup_queue = <- backup_queue_chan:
			if !backup_received && backup_queue.Cmd != 0 {

				if backup_queue.Floor != -1 {								//Add to buffer if valid order
					fmt.Println("I received: ", backup_queue)
					buffer[0][index_buffer] = backup_queue
					index_buffer += 1
				} else {
						for i := 0; i < index_buffer; i++ { //if there are no more valid orders, send on distribute channel
							distr_order_chan <- buffer[0][i]
						}
					index_buffer = 0
					backup_received = true  // after all orders are distributed
				}
			}

		case backup_id := <- backup_req_chan:
			if backup_id != config.LocalID {
				backup_queue := queue.RetrieveQueue()

				for i := 0; i < queue.Queue_size; i++ {
						backup_queue[backup_id][i].Cmd = config.OrdrAdd		//only send orders for the requested id

						transmit_backup_chan <- backup_queue[backup_id][i]
						fmt.Println("Sending:", backup_queue[backup_id][i], "to", backup_id)
						time.Sleep(100*time.Millisecond) // allahu akbar!
				}
			}
		case <- admit_loneliness.C:
			fmt.Println("No backup recived")
			admit_loneliness.Stop()
			/*
			var safety_queue [config.Num_elevs][queue.Queue_size]config.OrderStruct
			for i := 0 ; i < elevio.Num_floors ; i++{
				safety_queue[config.LocalID][i].Button = config.BT_Cab
				safety_queue[config.LocalID][i].Floor = i
				safety_queue[config.LocalID][i].ElevID = config.LocalID
				safety_queue[config.LocalID][i].Cmd = config.OrdrAdd
				safety_queue[config.LocalID][i].Timestamp = time.Now()
				distr_order_chan <- safety_queue[config.LocalID][i]
				//time.Sleep(1*time*Millisecond)
			}
			*/
			backup_received = true

		default:
			if !backup_received {
				request_backup_chan <-config.LocalID
			}
			time.Sleep(500*time.Millisecond)

		}

	}
}
