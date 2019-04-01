Elevator Project
================
Software for controlling `n` elevators working in parallel across `m` floors.

###Key design points:
 - UDP broadcast: for sending and receiving orders 
 - Peer-to-peer: communication for keeping track of the amount of nodes on the network. 
 - Watchdog: handling orders which are not confirmed executed. 
 - Queue: Storing orders for all nodes in a queue at each node. 
 - Backup: Nodes receive backup from other active nodes at initialization.


The software is fault tolerant in accordance with the given specifications.
  - Each node assume that the other nodes on the network are able to receive orders. The watchdog retransmit orders if they are not completed within a certain time limit. This ensures that no orders are left uncompleted. 

  - When receiving an order, the button lights are not activated unless at least the node knows that at least one other elevator has received that order.

The software consists of the following modules:
- **backupModule*
  - The backup module handles the backup functionality of the software. It includes a transmitting function and a receiving function. The node will at initialization request backup from other active peers, if any.
  - Interacts with:
    - **Network module** for transmitting and requesting backup.
    - **Queue module** for retrieving queue when transmitting and receiving backup.
  - Self-written code.

- **configPackage**
  - Collection of global types and constants used across the modules.
  - Included by close to all modules that use any globally used types or constants.
  - Self-written code.
 
- **driverModule**
  - Low level functionality for polling sensor inputs and setting outputs. 
  - Some self-written code, mostly pre-written.

- **elevsmModule**
  - Elevator State Machine Module.
  - Implementation of a finite state machine for a single elevator.
  - Interacts with driver module in order to start and stop elevator according to state.
  - Self-written code.

- **functionalityModule**.
  - Provides main functionality of the elevator at a higher abstraction. 
  - Interacts with:
    - **Elevator State machine module** for elevator state retrieval and command-sending.
    - **Queue module** for transmitting order completion and order execution.
    - **Driver module** for sensor inputs.
  - Self-written code.

- **networkModule**
  - Includes functionalities for communicating over a network such as
    UDP broadcast and peer-to-peer communication.
  - All modules transmitting or receiving from network interact with the network module.
  - Some self-written code, mostly pre-written.

- **queueModule**
  - Receiving and distributing orders and storing them in a queue. When a node receives a new order it is stored in its local queue and transmitted to the other nodes on the network.
  - **Watchdog** to handle uncompleted orders.
  - Interacts with:
    - **Driver module** for turning button lights on/off.
    - **Elevator state machine module** for state-retrieval.
    - **Functionality module**. See functionality module.
    - **Network module** for information about active peers and order transmitting/receiving.
  - Self-written code.