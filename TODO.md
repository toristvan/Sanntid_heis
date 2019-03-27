#TODO

## Code quality:

* Move constants and types to config.
* Try and workaround global variables
* Move goroutines out to main. Often transmit and receive, e.g. in orderdist()

## Queue:

* Is Queue() unnecessary?
* Work actual orderQueue into watchdog?
* Look at RemoveOrder() purpose of second for loop? Are we moving orders to beginning, is it necessary?
* Look at addToQueue(), is optimization logic necessary when we have stopArray()?

## Cost function:

* Works good enough
* Add functionality to take cab calls and order dir into cosnsideration
* Need access to stopArray, but also is to be called from queue... cyclic problem

## Backup:

* Enable request backup signal at startup/ e.g. ignored if all queues are empty
* Backup sends all orders and makes requesting elev add to queue
* Implement request backup upon going online after offline (Can make elev start in offline, and once gone online will request backup? 2-in-1 functionality)


## Issues:

* Program crashes if stays idle for too long... :( Queue becomes full?