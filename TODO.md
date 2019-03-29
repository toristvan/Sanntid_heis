#TODO

## Code quality:

* Move constants and types to config.
* Try and workaround global variables
* Move goroutines out to main. Often transmit and receive, e.g. in orderdist()

## Queue:

* Work actual orderQueue into watchdog?
* Look at RemoveOrder() purpose of second for loop? Are we moving orders to beginning (no), is it necessary?
* Look at addToQueue(), is optimization logic necessary when we have stopArray()?
*addToQueue(), remove input ID

## Cost function:

* Works good enough
* Add functionality to take cab calls and order dir into cosnsideration
* Need access to stopArray, but also is to be called from queue... cyclic problem
* Or can maybe use features of added queue.

## Backup:

* [DONE] Enable request backup signal at startup/ e.g. ignored if all queues are empty
* [DONE] Backup sends all orders and makes requesting elev add to queue
* [NOT DONE THIS] Implement request backup upon going online after offline (Can make elev start in offline, and once gone online will request backup? 2-in-1 functionality)



## Dead:

*Needs to tell others it's dead, so they'll take orders.

## Issues:

* Program crashes if stays idle for too long... :( Queue becomes full? Does not crash if being fed orders
* Frequent last prints before Crash:
* Idle
* Order confirmation Received
* Distribute order (often prints while crashed as well)

## Checklist refactoring
*FØKKING WORKED FOR MORE THAN 10 MINUTES FOR THE FIRST TIME, IT WAS WORTH 25+ HOURS ON THE LAB IN ONE SITTING
*Order distribution working (timer changed)
*Lots of gouroutines moved out
*Wakeup, fsm, etc with new timer...

ERRORS:
*Dead working (timer changed) [Dead becomes true several times - fix this]
Fix: Error in or in boolean statement
*Set button light error non-existen floor 255. When watchdog kicked in.

*Offline: Cab calls not working. Hall calls are.
Fix: Offline exception not added to distr order

RESULT:
Dead : retransmit doesn't happen, but watchdog handles retransim
Watchdog : Not triggering button_light error (as of yet?)
Offline : Works like a charm


