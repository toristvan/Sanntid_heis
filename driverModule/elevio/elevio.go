package elevio

import (
	"../../configPackage"
	"time"
	"sync"
	"net"
	"fmt"
)
/*---------------Using pre-written elevio package----------------*/
//Added functionality: 
// - Modified func Init(): make elev go to floor at initialization and turn of all lights.
// - func SetButtonLamp() and func SetFloorIndicator(): check boundaries before setting lights. 

const _pollRate = 50 * time.Millisecond

var _initialized bool = false
var _mtx sync.Mutex
var _conn net.Conn

func Init(addr string) { 
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true

	for i := 0; i < 3; i++ {
		for j := 0; j < config.Num_floors; j++ {
			SetButtonLamp(config.ButtonType(i), j, false)
		}
	}
	SetDoorOpenLamp(false)
	SetStopLamp(false)
	SetMotorDirection(config.MD_Down)
	for getFloor() == -1 {
		//wait until floor reached
	}
	SetMotorDirection(config.MD_Stop)
  	SetFloorIndicator(getFloor())

}

func SetMotorDirection(dir config.MotorDirection) {
	fmt.Println("Setting dir", dir)
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{1, byte(dir), 0, 0})
}

func SetButtonLamp(button config.ButtonType, floor int, value bool) {
	if (floor >=config.Ground_floor && floor < config.Num_floors) && (button >=config.BT_HallUp && button <= config.BT_Cab){
		_mtx.Lock()
		defer _mtx.Unlock()
		_conn.Write([]byte{2, byte(button), byte(floor), toByte(value)})
	}
}

func SetFloorIndicator(floor int) {
	if (floor >=config.Ground_floor && floor < config.Num_floors) {
		_mtx.Lock()
		defer _mtx.Unlock()
		_conn.Write([]byte{3, byte(floor), 0, 0})
	}
}

func SetDoorOpenLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{4, toByte(value), 0, 0})
}

func SetStopLamp(value bool) {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{5, toByte(value), 0, 0})
}

func PollButtons(receiver chan<- config.ButtonEvent) {
	prev := make([][3]bool, config.Num_floors)
	for {
		time.Sleep(_pollRate)
		for f := config.Ground_floor; f < config.Num_floors; f++ {
			for b := config.ButtonType(0); b < 3; b++ {
				v := getButton(b, f)
				if v != prev[f][b] && v != false {
					receiver <- config.ButtonEvent{f, config.ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := getFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := getObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

func getButton(button config.ButtonType, floor int) bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{6, byte(button), byte(floor), 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getFloor() int {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{7, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	if buf[1] != 0 {
		return int(buf[2])
	} else {
		return -1
	}
}

func getStop() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{8, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func getObstruction() bool {
	_mtx.Lock()
	defer _mtx.Unlock()
	_conn.Write([]byte{9, 0, 0, 0})
	var buf [4]byte
	_conn.Read(buf[:])
	return toBool(buf[1])
}

func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}
