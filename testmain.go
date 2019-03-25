package main

import (
    "./functionalityModule"
    "./queueModule"
    "./driverModule/elevio"
    "time"
)

func main() {
    queue.InitQueue()
    elevio.Init("localhost:15657") //, num_floors)
    go elevclient.ElevRunner()

    for {
        time.Sleep(5*time.Second)
    }

}