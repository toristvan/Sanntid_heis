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
    offline_chan := make (chan bool)
    
    initElevNode()
    elevio.Init(Sprintf("localhost:1000%d", config.LocalID)) //, num_floors)  //For simulators
    //elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators

    //elevstatus goroutines
    go elevclient.IsElevDead(is_dead_chan)

    go peers.CheckOffline(20003, offline_chan)

    //Tidligere init i queue/Queue
    go elevclient.ElevRunner(elev_cmd_chan, delete_order_chan )
    go queue.DistributeOrder(distr_order_chan, add_order_chan, delete_order_chan, offline_chan)
    go queue.ReceiveOrder(add_order_chan)

    //Tidligere Init i elevatorClient/ElevRunner
    go elevclient.IOwrapper(raw_order_chan)
    go queue.Queue(raw_order_chan, distr_order_chan, add_order_chan ,execute_chan)
    go fsm.ElevStateMachine(elev_cmd_chan)
    go elevclient.ExecuteOrder(execute_chan)

    //Init i main
    //go backUp()
    
    for {
        select{
        case deadness := <- is_dead_chan:
            switch deadness{
            case true:
                Printf("DEAD!\n")
            case false:
                Printf("ALIVE!\n")
            }

        }
		
		time.Sleep(1*time.Second)

    }
}


                                                                                                                                                                 



