package main

import (
	"./driverModule/elevio"
	"./queueModule"
	"./functionalityModule"
	)

func main() {

	queue.InitQueue()
	elevio.Init("localhost:15657") //, num_floors)
	go elevclient.RunElevator()
}
