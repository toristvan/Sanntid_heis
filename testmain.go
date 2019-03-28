package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./fsmModule"
    //"./networkModule/bcast"
    "./networkModule/peers"
    //"./backupModule"
    "time"
    ."fmt"
    //"os/exec"
    //"strconv"
)
/*
type networkOrderStruct struct{
    Order config.OrderStruct
    req_update bool
    send_update bool
}*/

func initElevNode() int {
    var id int

    Println("Set id")
    Scanf("%d", &id)

    for id > config.Num_elevs{
        Println("Invalid id! Shame on you")
        Println("Set id")
    	Scanf("%d", &id)
    }
    config.InitConfigData(id)
    queue.InitQueue()
    Println("Id set to",id,"number of elevators", config.Num_elevs)
    return id
}

//Works fine-ish. Needs to send ID and store orders in a queue.
//Needs to figure out properly when program is frozen. (When backup should take over)
//Sometimes spawns unexpectedly...
func backUp(id int, checkbackup_chan chan<- bool){
    /*
    //backUpCmd := exec.Command("cmd.exe","/C","start", `c:\Users\torge\Documents\Go-code\project-gruppe-61-real\testmain.go"`)
    //var backup_id int
    
    //For Linux:
    backUpCmd := exec.Command("gnome-terminal", "-x", "go", "run", "/home/student/Desktop/GR61REAL/project-gruppe-61-real/testmain.go")
    //For Windows:
   // backUpCmd := exec.Command("cmd.exe","/C","start") //For Windows. Åpner et nytt cmd vindu men, har ikke klart å få til å kjøre et nytt go script

    var backup_id int
    primary_id := strconv.Itoa(id) //from int to string
    primary := false

    backup_receive_chan := make(chan peers.PeerUpdate)
    offline_chan := make(chan bool)

    go peers.Transmitter(config.Backup_port, primary_id, offline_chan)
    go peers.CheckOffline(config.Backup_port, offline_chan)
    //go bcast.Receiver(config.Backup_port, backup_receive_chan)
    go peers.Receiver(config.Backup_port, backup_receive_chan)

    for{
    	timeoutTicker := time.NewTicker(2000*time.Millisecond)
        defer timeoutTicker.Stop()

        select{
        case tmp := <- backup_receive_chan:
          backup_id, _ = strconv.Atoi(tmp.New) //From string to int
          Println("backup_id",backup_id)
        case <- timeoutTicker.C:
        	if !primary {
        		//offline_chan <- true //This will block forever and produces hard to find bugs
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
    */
    checkbackup_chan <- true

}



func main() {
    //Queue channels
    //add_order_chan := make(chan config.OrderStruct)
    delete_order_chan := make(chan config.OrderStruct)
    is_dead_chan := make (chan bool)
    retransmit_last_order_chan := make (chan bool)
    //backup_queue_chan := make(chan int)
    //transmit_backup_chan := make(chan [config.Num_elevs][10]config.OrderStruct)
    //backup_req_chan := make(chan int)

    //elevrunner channels
    //raw_order_chan   := make (chan config.OrderStruct)
    distr_order_chan := make(chan config.OrderStruct)
    execute_chan  := make (chan config.OrderStruct)
    elev_cmd_chan := make (chan config.ElevCommand)

    checkbackup_chan := make(chan bool)
    offline_chan := make (chan bool)

    id := initElevNode()

    go backUp(id, checkbackup_chan)


    //elevio.Init(Sprintf("localhost:2000%d", config.LocalID)) //, num_floors)  //For simulators
    elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators


    //ny_queue :=  queue.RetrieveQueue()
    //backup_queue_chan <- ny_queue

    //Mainloop only runs after backup-check fails (No connection with primary function)
    //Can run without backup if backUp is commented out and only sends bool through checkbackup_chan
    Printf("\n\n-------------INITIALIZED-------------\n")
    for {
        select{
        case <- checkbackup_chan:
          //debugging purposes
          //go queue.PrintQueue()
  		    go queue.DistributeOrder(distr_order_chan, execute_chan, delete_order_chan, offline_chan, retransmit_last_order_chan)
  		    go queue.ReceiveOrder(execute_chan, is_dead_chan, retransmit_last_order_chan)
  		    go elevclient.ExecuteOrder(execute_chan)


          go elevclient.IsElevDead(is_dead_chan)
          go peers.CheckOffline(20003, offline_chan)

          //Backup

          //Println("Primary")
          //Tidligere init i queue/Queue
          go elevclient.ElevRunner(elev_cmd_chan, delete_order_chan )

          //Tidligere Init i elevatorClient/ElevRunner
          go elevclient.IOwrapper(distr_order_chan)
          //go queue.Queue(/*raw_order_chan,*/ distr_order_chan, add_order_chan ,execute_chan)
          go fsm.ElevStateMachine(elev_cmd_chan)

          //go bcast.Receiver(config.Backup_port, backup_req_chan)
          //go bcast.Transmitter(config.Backup_port, transmit_backup_chan)
          //go backup.RequestBackup(distr_order_chan, backup_req_chan, transmit_backup_chan)
          //Println(queue.RetrieveQueue())
		default:
			time.Sleep(1*time.Second)
		}
    }
}
