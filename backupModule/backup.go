package backup

import(
	"../networkModule/bcast"
	"../queueModule"
	"../configPackage"
	"../driverModule/elevio" //remove when num_floors are removed
	"fmt"
	"time"

)


func RequestBackup(distr_order_chan chan<- config.OrderStruct, backup_req_chan <-chan int, transmit_backup_chan chan<- [config.Num_elevs][queue.Queue_size]config.OrderStruct/*, _queue_chan chan<- [config.Num_elevs][queue.Queue_size]config.OrderStruct*/) {
	var backup_queue [config.Num_elevs][queue.Queue_size]config.OrderStruct
	backup_received := false

	backup_queue_chan := make(chan [config.Num_elevs][queue.Queue_size]config.OrderStruct)
	request_backup_chan := make(chan int)
	
	//Transmit backup request and receive backup-queue
	go bcast.Receiver(config.Backup_port, backup_queue_chan)
	go bcast.Transmitter(config.Backup_port, request_backup_chan)

	request_backup_chan <- config.LocalID

	admit_loneliness := time.NewTicker(5*time.Second)

	for{
		select{
		case backup_queue = <- backup_queue_chan:
			if !backup_received && backup_queue[0][0].Cmd != 0  {  //No valid_order.Cmd in Queue should be 0
				fmt.Println("I received: ", backup_queue)
				for j := 0; j < config.Num_elevs; j++ {
					for i := 0; i < queue.Queue_size; i++ {
						distr_order_chan <- backup_queue[i][j]
						//time.Sleep(100*time.Millisecond)
					}
				}
				backup_received = true
			} else {
			fmt.Println("I received something wrong. Cmd:", backup_queue[0][0])
			}
		case backup_id :=<- backup_req_chan:
			if backup_id != config.LocalID {
				backup_queue := queue.RetrieveQueue()
				fmt.Println("Sending:", backup_queue)
				transmit_backup_chan <- backup_queue
				time.Sleep(1*time.Second)
			}	
		case <- admit_loneliness.C:
			admit_loneliness.Stop()
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
			backup_received = true
		default:
			if !backup_received {
				request_backup_chan <-config.LocalID
			}
			time.Sleep(500*time.Millisecond) 
		}

	}
}