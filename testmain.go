package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./networkModule/bcast"
    "time"
    "os/exec"
    ."fmt"
)
var primary bool = false

type networkOrderStruct struct{
    Order config.OrderStruct
    req_update bool
    send_update bool
}

//Ser ut til at man får samme problem som sånn med IO.
//Nå er broacast funksjonaliteten satt en overordnet funksjon
//Fungerer ganske bra. Eneste ulempen er at kanalene for å sending og mottakning får en sabla lang vei å gå 
//Som det er implementert nå går kanalbanen: broadCastHub -> Elevrunner -> Queue -> DistributeOrder (noe som er litt jalla)

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

func initElevNode(){
    var num_of_elev int = 3
    var id int

    Println("Set id")
    Scanf("%d", &id)

    if id > num_of_elev{
        Println("Invalid id! Shame on you")
        id = 0
    } 

    config.InitConfigData(id, num_of_elev)
    queue.InitDataQueue()
    Println("Id set to",id,"number of elevators", num_of_elev)
}


func backUp(){
    backUpCmd := exec.Command("gnome-terminal", "-x", "go", "run", "/home/student/GR61REAL/project-gruppe-61-real/testmain.go")

    for{
        if primary {
        err := backUpCmd.Run()

            if err != nil {
                Println(err)
            }
        }

        timeoutTicker := time.NewTicker(3000*time.Millisecond)
        defer timeoutTicker.Stop()

        select{
        case <- timeoutTicker.C:
            primary = true
        }
    }
}



func main() {
    var reqUpdateSwitch bool 

    rec_sink_chan       := make(chan config.OrderStruct,10)
    trans_source_chan   := make(chan config.OrderStruct,10)

    rec_main_chan       := make(chan config.OrderStruct)
    trans_main_chan     := make(chan config.OrderStruct)
    req_update_queue    := make(chan bool)

    Offline_notify_chan := make(chan bool)



    initElevNode()

    elevio.Init("localhost:15657") //, num_floors)
    go elevclient.ElevRunner(trans_source_chan, rec_sink_chan)
    go broadCastHub(rec_main_chan, trans_main_chan, req_update_queue, Offline_notify_chan)
    //go backUp()

    for {
        select{
        case tmp := <- trans_source_chan: //How deep does the rabbit hole go?
            Println("tx main", tmp)
            trans_main_chan <- tmp

        case tmp := <- rec_main_chan:
            Println("rx main",tmp)
            rec_sink_chan  <- tmp
                
        case tmp := <- Offline_notify_chan:
            if tmp {
                reqUpdateSwitch = true
                Println("offline", reqUpdateSwitch)
            }else if reqUpdateSwitch {
                req_update_queue <- true
            }else{
                req_update_queue <- false
            }
        }
    }
}


                                                                                                                                                                 



