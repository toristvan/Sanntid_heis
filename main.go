package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./elevsmModule"
    "./networkModule/bcast"
    "./networkModule/peers"
    "./backupModule"
    ."fmt"
    "strconv"
)

func initElevNode() {
    var id int

    Println("Set id")
    Scanf("%d", &id)

    for id > config.Num_elevs{
        Println("Invalid id! must be from 0-", config.Num_elevs-1)
        Println("Set id")
    	Scanf("%d", &id)
    }
    config.InitConfigData(id)
    queue.InitQueue()
    Println("ID set to: ",id,"\n Number of elevators: ", config.Num_elevs)
}

func main() {

    //Peers channels
    peers_update_chan           := make (chan peers.PeerUpdate)
    peer_tx_enable_chan         := make (chan bool)

    //Queue channels
    rec_order_chan              := make (chan config.OrderStruct)
    trans_conf_chan             := make (chan config.OrderStruct)
    trans_order_chan            := make (chan config.OrderStruct)
    distr_order_chan            := make(chan config.OrderStruct)
    delete_order_chan           := make(chan config.OrderStruct)
    retransmit_last_order_chan  := make (chan bool)
    execute_chan                := make (chan config.OrderStruct)
    
    //Elevstatus channels
    is_dead_chan                := make (chan bool)
    offline_chan                := make (chan bool)

    //elevrunner channels
    elev_cmd_chan               := make (chan config.ElevCommand)
    wakeup_chan                 := make (chan bool)

    //backup channels
    transmit_backup_chan        := make(chan config.OrderStruct)
    received_backup_request_chan:= make(chan int)

    //io channels
    drv_floors_dead_chan        := make (chan int)
    drv_floors_run_chan         := make (chan int)
    drv_buttons_chan            := make(chan config.ButtonEvent)

    initElevNode()
    Printf("\n\n-------------INITIALIZING-------------\n")

    //elevio.Init(Sprintf("localhost:2000%d", config.Local_ID)) //, num_floors)  //For simulators
    elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators

    Printf("CHECKING FOR BACKUP...")

    //Peers goroutines
    go peers.Transmitter(config.Peer_port, strconv.Itoa(config.Local_ID), peer_tx_enable_chan)
    go peers.Receiver(config.Peer_port, peers_update_chan)
    go peers.CheckForPeers(peers_update_chan)
    
    //IO goroutines
    go elevio.PollFloorSensor(drv_floors_run_chan)
    go elevio.PollFloorSensor(drv_floors_dead_chan)
    go elevio.PollButtons(drv_buttons_chan)
    go elevopr.IOwrapper(distr_order_chan, drv_buttons_chan)

    //Queue goroutines
    go bcast.Receiver(config.Order_port, rec_order_chan)
    go bcast.Transmitter(config.Order_port, trans_conf_chan)  
    go bcast.Transmitter(config.Order_port, trans_order_chan)
    go queue.DistributeOrder(distr_order_chan, execute_chan, delete_order_chan, offline_chan, retransmit_last_order_chan, trans_order_chan)
    go queue.ReceiveOrder(execute_chan, is_dead_chan, retransmit_last_order_chan, rec_order_chan, trans_conf_chan)
    go queue.Watchdog(distr_order_chan)
    go elevopr.ExecuteOrder(execute_chan)

    //Elevstatus goroutines
    go peers.CheckOffline(config.Offline_port, offline_chan)
    go elevopr.IsElevDead(is_dead_chan, drv_floors_dead_chan)

    //Running goroutines
    go elevopr.ElevOperator(elev_cmd_chan, delete_order_chan, wakeup_chan, drv_floors_run_chan)
    go elevopr.ElevWakeUp(wakeup_chan)
    go elevsm.ElevStateMachine(elev_cmd_chan)

    //Backup goroutines
    go bcast.Receiver(config.Backup_port, received_backup_request_chan)
    go bcast.Transmitter(config.Backup_port, transmit_backup_chan)
    go backup.RequestBackup(distr_order_chan)
    go backup.TransmitBackup(received_backup_request_chan, transmit_backup_chan)

    select{

    }
}