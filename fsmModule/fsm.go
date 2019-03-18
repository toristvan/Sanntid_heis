package fsm

import(
  ."fmt"
  "time"
)

// "<-chan" receive
// "chan<-" send

type ElevState int
var state ElevState

const (
	idle 			ElevState = 1
	Moving 		ElevState = 2
	goingUp 	ElevState = 3
	goingDown ElevState = 4
	atFloor 	ElevState = 5
)

func eventHandler(someEvent <-chan int, nextState chan<- ElevState) {

  event := <- someEvent
  switch state {
  case idle:

    switch event {
    case 1:   nextState <- idle
    case 2:   nextState <- Moving
    case 5:   nextState <- atFloor
    default:  nextState <- idle
    }
    Println("idle:")
    time.Sleep(1000 * time.Millisecond)

  case Moving:

    switch event {
    case 3:   nextState <- goingUp
    case 4:   nextState <- goingDown
    default:  nextState <- Moving
    }
    Println("Moving:")
    time.Sleep(1000 * time.Millisecond)

  case goingUp:

    if event == 5 { nextState <- atFloor
    }else{          nextState <- goingUp}
    Println("goingUp:")
    time.Sleep(1000 * time.Millisecond)

  case goingDown:

    if event == 5 { nextState <- atFloor
    }else{          nextState <- goingDown}
    Println("goingDown:")
    time.Sleep(1000 * time.Millisecond)

  case atFloor:

    if event == 6 { nextState <- idle
    }else{          nextState <- atFloor}
    Println("atFloor:")
    time.Sleep(1000 * time.Millisecond)

  }
}

func selectState(sensorInput int, event chan<- int, nextState <-chan ElevState) {

  event <- sensorInput
  select{
  case switchEvent := <- nextState:
      switch switchEvent{
      case 1:   state = idle
      case 2:   state = Moving
      case 3:   state = goingUp
      case 4:   state = goingDown
      case 5:   state = atFloor
      }
  }
}

func main() {
  var input int
  state = idle
  nextState := make(chan ElevState, 1)
  event := make(chan int, 1)

  for{
    Scanf("%d", &input)
    go selectState(input, event, nextState)
    go eventHandler(event, nextState)
  }
}
