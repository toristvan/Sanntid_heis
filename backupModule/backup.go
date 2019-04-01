package backup

import(
	"../networkModule/bcast"
	"../queueModule"
	"../configPackage"
	"fmt"
	"time"

)

//Consists of both receiving and transmitting end. Channel to ask for backup is closed once backup is received.
//The channel will trigger transmitting end to transmit backup to ID requesting it.
func RequestBackup(distr_order_chan chan<- config.OrderStruct) {
	var backup_received bool = false
	var buffer [18]config.OrderStruct
	var index_buffer int = 0
	var in_buffer bool = false

	receive_backup_chan := make(chan config.OrderStruct)
	request_backup_chan := make(chan int)

	//Transmit backup request and receive backup-queue
	go bcast.Receiver(config.Backup_port, receive_backup_chan)
	go bcast.Transmitter(config.Backup_port, request_backup_chan)
	request_backup_chan <- config.Local_ID
	admit_loneliness := time.NewTicker(5*time.Second)

	for{
		select{
		case backup_order := <- receive_backup_chan:
			defer close(receive_backup_chan)
			defer close(request_backup_chan)
			admit_loneliness.Stop()
			if !backup_received {
				for i:= 0 ; i < index_buffer; i++ {
					if buffer[i].ElevID == backup_order.ElevID && buffer[i].Button == backup_order.Button && buffer[i].Floor == backup_order.Floor {
						in_buffer = true
						break
					}
				}
				if !in_buffer {
					buffer[index_buffer] = backup_order
				}
				in_buffer = false
				index_buffer += 1
				//If buffer full	(Size = Maximum amount of unique orders)
				if index_buffer >= (config.Num_elevs + 2)*config.Num_floors -2 {
						for i := 0; i < index_buffer; i++ { 
							distr_order_chan <- buffer[i]
						}
					//All backup orders received
					backup_received = true  
					fmt.Println("\nBACKUP RECEIVED\n")
    				fmt.Printf("\n\n-------------INITIALIZED-------------\n")
				}
			}
		//If too long without receiving timeout
		case <- admit_loneliness.C:
			fmt.Println("\nNO BACKUP RECEIVED\n")
    		fmt.Printf("\n\n-------------INITIALIZED-------------\n")
			admit_loneliness.Stop()
			backup_received = true
		case <- time.After(50*time.Millisecond):
			if !backup_received {
				request_backup_chan <-config.Local_ID
			}
		}
	}
}

func TransmitBackup(received_backup_request_chan <-chan int, transmit_backup_chan chan<- config.OrderStruct) {
	for {
		backup_id := <- received_backup_request_chan:
		if backup_id != config.Local_ID {
			backup_send_queue := queue.RetrieveQueue()
			for i := 0; i < config.Num_elevs; i++ {
				for j := 0; j < config.Queue_size; j++ {
					backup_send_queue[i][j].Cmd = config.OrdrAdd
					//Add to buffer if valid order
					if backup_send_queue[i][j].Floor != -1 {
						transmit_backup_chan <- backup_send_queue[i][j]
					}
				}
			}
		}
	}
}