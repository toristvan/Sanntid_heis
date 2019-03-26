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

type networkOrderStruct struct{
    Order config.OrderStruct
    req_update bool
    send_update bool
}

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

//Works fine-ish. Needs to send ID and store orders in a queue.
//Needs to figure out properly when program is frozen. (When backup should take over)
//Sometimes spawns unexpectedly...
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
    is_dead_chan := make (chan bool)


	//elevrunner channels
  	raw_order_chan   := make (chan config.OrderStruct)  
  	execute_chan  := make (chan config.OrderStruct)    
  	elev_cmd_chan := make (chan config.ElevCommand)

  	checkbackup_chan := make(chan bool)
    offline_chan := make (chan bool)
    
    initElevNode()
    go backUp(checkbackup_chan)
    
    elevio.Init(Sprintf("localhost:2000%d", config.LocalID)) //, num_floors)  //For simulators
    //elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators
    




    //Mainloop only runs after backup-check fails (No connection with primary function)
    //Can run without backup if backUp is commented out and only sends bool through checkbackup_chan
    for {
        select{
        
		case <- checkbackup_chan: 
			Println("Primary")
		    //Tidligere init i queue/Queue
		    go elevclient.ElevRunner(elev_cmd_chan, delete_order_chan )
		    go queue.DistributeOrder(distr_order_chan, add_order_chan, delete_order_chan, offline_chan)
		    go queue.ReceiveOrder(add_order_chan, is_dead_chan)

		  
    		go elevclient.IsElevDead(is_dead_chan)
		    go peers.CheckOffline(20003, offline_chan)

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


                                                                                                                                                                 



