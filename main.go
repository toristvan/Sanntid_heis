package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./fsmModule"
    "./networkModule/bcast"
    "./networkModule/peers"
    "./backupModule"
    //"time"
    ."fmt"
    //"os/exec"
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
    //backup_queue_chan := make(chan int)
    
    //Elevstatus channels
    is_dead_chan                := make (chan bool)
    offline_chan                := make (chan bool)

    //elevrunner channels
    elev_cmd_chan               := make (chan config.ElevCommand)
    wakeup_chan                 := make (chan bool)

    //transmit_backup_chan := make(chan [config.Num_elevs][10]config.OrderStruct)
    transmit_backup_chan        := make(chan config.OrderStruct)
    backup_req_chan             := make(chan int)

    //checkbackup_chan := make(chan bool)
    drv_floors_dead_chan        := make (chan int)
    drv_floors_run_chan         := make (chan int)
    drv_buttons_chan            := make(chan config.ButtonEvent)//moved

    //deadlock channel
    deadlock_chan               := make (chan bool)

    initElevNode()
    Printf("\n\n-------------INITIALIZING-------------\n")

    //go backUp(id, checkbackup_chan)
    elevio.Init(Sprintf("localhost:2000%d", config.LocalID)) //, num_floors)  //For simulators
    //elevio.Init(Sprintf("localhost:15657"))//, num_floors)                      //For elevators

    Printf("CHECKING FOR BACKUP...")

    //Peers goroutines
    go peers.Transmitter(config.Peer_port, strconv.Itoa(config.LocalID), peer_tx_enable_chan)
    go peers.Receiver(config.Peer_port, peers_update_chan)
    go peers.CheckForPeers(peers_update_chan)
    
    //IO goroutines
    go elevio.PollFloorSensor(drv_floors_run_chan)
    go elevio.PollFloorSensor(drv_floors_dead_chan)
    go elevio.PollButtons(drv_buttons_chan)//moved
    go elevclient.IOwrapper(distr_order_chan, drv_buttons_chan)

    //Queue goroutines
    go bcast.Receiver(config.Order_port, rec_order_chan)
    go bcast.Transmitter(config.Order_port, trans_conf_chan)  

    go bcast.Transmitter(config.Order_port, trans_order_chan)
    go queue.DistributeOrder(distr_order_chan, execute_chan, delete_order_chan, offline_chan, retransmit_last_order_chan, trans_order_chan)
    go queue.ReceiveOrder(execute_chan, is_dead_chan, retransmit_last_order_chan, rec_order_chan, trans_conf_chan)
    go queue.Watchdog(distr_order_chan)
    go elevclient.ExecuteOrder(execute_chan)

    //Elevstatus goroutines
    go elevclient.IsElevDead(is_dead_chan, drv_floors_dead_chan)
    go peers.CheckOffline(config.Offline_port, offline_chan)
    go elevclient.ElevRunner(elev_cmd_chan, delete_order_chan, wakeup_chan, drv_floors_run_chan)


    //Running goroutines
    go elevclient.ElevWakeUp(wakeup_chan)
    go fsm.ElevStateMachine(elev_cmd_chan)

    //Backup goroutines
    go bcast.Receiver(config.Backup_port, backup_req_chan)
    go bcast.Transmitter(config.Backup_port, transmit_backup_chan)
    go backup.RequestBackup(distr_order_chan, backup_req_chan, transmit_backup_chan)

    select{
      case <- deadlock_chan:
    }
}