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
	var backup_received bool = false
	var buffer [18]config.OrderStruct
	var index_buffer int = 0
	var in_queue bool = false

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
		case backup_order := <- backup_queue_chan:
			defer close(backup_queue_chan)
			defer close(request_backup_chan)
			admit_loneliness.Stop()
			if !backup_received {
				for i:= 0 ; i < index_buffer; i++ {
					if buffer[i].ElevID == backup_order.ElevID && buffer[i].Button == backup_order.Button && buffer[i].Floor == backup_order.Floor {
						//fmt.Println(buffer[i].Button, backup_order.Button, buffer[i].Floor , backup_order.Floor)
						in_queue = true
						break
					}
				}
				if !in_queue {
					buffer[index_buffer] = backup_order
					fmt.Println("I received: ", buffer[index_buffer])
				}
				in_queue = false
				index_buffer += 1
				
				if index_buffer >= 18 {
						for i := 0; i < index_buffer; i++ { //if there are no more valid orders, send on distribute channel
							distr_order_chan <- buffer[i]
						}
					backup_received = true  // after all orders are distributed
					fmt.Println("\nBACKUP RECEIVED\n")
    				fmt.Printf("\n\n-------------INITIALIZED-------------\n")

				}
			}

		case backup_id := <- backup_req_chan:
			if backup_id != config.LocalID {
				backup_send_queue := queue.RetrieveQueue()
				for i := 0; i < config.Num_elevs; i++ {
					for j := 0; j < queue.Queue_size; j++ {
						backup_send_queue[i][j].Cmd = config.OrdrAdd		//only send orders for the requested id

						if backup_send_queue[i][j].Floor != -1 {								//Add to buffer if valid order
							transmit_backup_chan <- backup_send_queue[i][j]
							fmt.Println("Sending:", backup_send_queue[i][j], "to", backup_id)
							//time.Sleep(20*time.Millisecond) // allahu akbar!
						}
					}
				}
			}
		case <- admit_loneliness.C:
			fmt.Println("\nNO BACKUP RECEIVED\n")
    		fmt.Printf("\n\n-------------INITIALIZED-------------\n")
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

		case <- time.After(50*time.Millisecond):
			if !backup_received {
				request_backup_chan <-config.LocalID
			}
		}
	}
}
