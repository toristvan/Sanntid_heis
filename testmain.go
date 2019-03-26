package main

import (
    "./configPackage"
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "./networkModule/bcast"
    //"time"
    ."fmt"
)

//Ser ut til at man får samme problem som sånn med IO.
//Nå er broacast funksjonaliteten satt en overordnet funksjon
//Fungerer ganske bra. Eneste ulempen er at kanalene for å sending og mottakning får en sabla lang vei å gå 
//Som det er implementert nå går kanalbanen: broadCastHub -> Elevrunner -> Queue -> DistributeOrder (noe som er litt jalla)

func broadCastHub(recive_chan  chan <- config.OrderStruct, transmit_chan <-chan config.OrderStruct, Offline_notify_chan chan<- bool){
    var port int        = 20007

    trans         := make (chan config.OrderStruct)
    rec           := make (chan config.OrderStruct)
    offline_alert := make (chan bool)

    go bcast.Receiver(port,rec)
    go bcast.Transmitter(port,offline_alert, trans)

    for{
        select{
            case distribute_rec := <-rec:
                recive_chan <- distribute_rec 

            case distribute_trans := <- transmit_chan:
                //Println("Transmitting order")
                trans <- distribute_trans

            case is_offline := <-offline_alert: 
            if is_offline {
                Offline_notify_chan <- is_offline
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

func main() {
    initElevNode()


    rec_main_chan       := make(chan config.OrderStruct,10)
    trans_main_chan     := make(chan config.OrderStruct,10)

    recive_chan         := make(chan config.OrderStruct)
    transmit_chan       := make(chan config.OrderStruct)

    Offline_notify_chan := make(chan bool)

    //queue.InitQueue()
    elevio.Init("localhost:15657") //, num_floors)
    go elevclient.ElevRunner(trans_main_chan, rec_main_chan)
    go broadCastHub(recive_chan, transmit_chan, Offline_notify_chan)
    //go test()

    for {
        select{
        case tmp := <- trans_main_chan: //How deep does the rabbit hole go?
            Println("tx main", tmp)
            transmit_chan <- tmp

        case tmp := <- recive_chan:
            Println("rx main",tmp)
            rec_main_chan <- tmp

        case tmp := <-Offline_notify_chan:
            if tmp {
                Println("offline")
            }
        }


        //time.Sleep(200*time.Second)
    }

}