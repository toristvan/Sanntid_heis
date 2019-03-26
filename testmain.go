package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./fsmModule"
    "./networkModule/bcast"
    "./networkModule/peers"
    "time"
    "os/exec"
    ."fmt"
)
//var primary bool = false

type networkOrderStruct struct{
    Order config.OrderStruct
    req_update bool
    send_update bool
}

//Ser ut til at man får samme problem som sånn med IO.
//Nå er broacast funksjonaliteten satt en overordnet funksjon
//Fungerer ganske bra. Eneste ulempen er at kanalene for å sending og mottakning får en sabla lang vei å gå 
//Som det er implementert nå går kanalbanen: broadCastHub -> Elevrunner -> Queue -> DistributeOrder (noe som er litt jalla)
/*
func broadCastHub(recive_chan  chan <- config.OrderStruct, transmit_chan <-chan config.OrderStruct, req_update_queue <- chan bool, Offline_notify_chan chan<- bool){
    var port int = 20007
    var txPacket networkOrderStruct

    trans         := make (chan networkOrderStruct)
    rec           := make (chan networkOrderStruct)
    offline_alert := make (chan bool)

    go bcast.Receiver(port,rec)
    go bcast.Transmitter(port, offline_alert, trans)

    for{
        select{
            case distribute_rec := <-rec:
                recive_chan <- distribute_rec.Order
                if distribute_rec.req_update {

                    var queue = queue.RetriveQueue()

                    txPacket.Order = queue[0][0]

                    Println(txPacket.Order)
                    txPacket.req_update = false
                    trans <- txPacket
                }

            case distribute_trans := <- transmit_chan:
                Println("Transmitting order")
                txPacket.Order = distribute_trans
                trans <- txPacket

            case is_offline := <-offline_alert: 
                if is_offline {
                    Offline_notify_chan <- is_offline
                }

            case tmp := <- req_update_queue:
                if tmp {
                    txPacket.req_update = true
                    trans <- txPacket
                }
        }
    }
}
*/

func initElevNode(){
    var num_of_elev int = 3
    var id int

    Println("Set id")
    Scanf("%d", &id)

    for id > num_of_elev{
        Println("Invalid id! Shame on you")
        Println("Set id")
    	Scanf("%d", &id)
    } 

    config.InitConfigData(id, num_of_elev)
    queue.InitQueue()
    Println("Id set to",id,"number of elevators", num_of_elev)
}


func backUp(checkbackup_chan chan<- bool){
    backUpCmd := exec.Command("gnome-terminal", "-x", "go", "run", "/home/student/Desktop/GR61REAL/project-gruppe-61-real/testmain.go")

    Println("backup init")
    primary := false

    backup_receive_chan := make(chan bool)
    offline_chan := make(chan bool)

    go peers.Transmitter(config.Backup_port, "Tester", offline_chan)
    go bcast.Receiver(config.Backup_port, backup_receive_chan)

    for{
    	timeoutTicker := time.NewTicker(1000*time.Millisecond)
        defer timeoutTicker.Stop()

        select{
        case <- backup_receive_chan:
        case <- timeoutTicker.C:
        	if !primary {
        		offline_chan <- true
        		select {
        		case offline_check := <- offline_chan:
        			if !offline_check {
			            err := backUpCmd.Run()

			            if err != nil {
			                Println(err)
			            }
			            primary = true
			            checkbackup_chan <- true
		        	} else {
	        			Println("Offline, no spawn")
		        	}
		        }
	        }
        }
    }
}



func main() {
    //Queue channels
    add_order_chan := make(chan config.OrderStruct) 
	distr_order_chan := make(chan config.OrderStruct)  
	delete_order_chan := make(chan config.OrderStruct)

	//elevrunner channels
  	raw_order_chan   := make (chan config.OrderStruct)  
  	execute_chan  := make (chan config.OrderStruct)    
  	elev_cmd_chan := make (chan config.ElevCommand)

  	checkbackup_chan := make(chan bool)
    
    initElevNode()
    go backUp(checkbackup_chan)
    
    elevio.Init(Sprintf("localhost:2000%d", config.LocalID)) //, num_floors)  //For simulators
    //elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators


    //Init i main
    
    for {
		
		select {
		case <- checkbackup_chan: 
			Println("Primary")
		    //Tidligere init i queue/Queue
		    go elevclient.ElevRunner(elev_cmd_chan, delete_order_chan )
		    go queue.DistributeOrder(distr_order_chan, add_order_chan, delete_order_chan)
		    go queue.ReceiveOrder(add_order_chan)

		    //Tidligere Init i elevatorClient/ElevRunner
		    go elevclient.IOwrapper(raw_order_chan)
		    go queue.Queue(raw_order_chan, distr_order_chan, add_order_chan ,execute_chan)
		    go fsm.ElevStateMachine(elev_cmd_chan)
		    go elevclient.ExecuteOrder(execute_chan)
		default:
			time.Sleep(1*time.Second)
		}
    }
}


                                                                                                                                                                 



