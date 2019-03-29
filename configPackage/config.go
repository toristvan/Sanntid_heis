package config

import (
	"fmt"
	"time"
)
const MaxCost int = 10 // Num_floors +2
const Num_elevs int = 3
//const Num_floors int = 4

var LocalID int
var Order_port int
var Backup_port	int


var Current_floor int

func InitConfigData(id int){
	LocalID   = id
	Backup_port	= 10070 //+ LocalID
	Order_port = 10005
	fmt.Println("Configuration data initialized")
}

type ElevStateType int
const (
	Idle		ElevStateType = 0
	GoingUp 				= 1
	GoingDown 	 			= 2
	AtFloor 		 		= 3
)

type MotorDirection int
const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int
const (
	BT_Invalid	 ButtonType	= -1
	BT_HallUp				= 0
	BT_HallDown            	= 1
	BT_Cab                 	= 2
)

type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

type OrderStruct struct
{
	Button ButtonType
	Floor int
	Timestamp time.Time
	Cost int
	Cmd OrderCommand
	ElevID int
	MasterID int
	SenderID int
}

type OrderCommand int
const (
	OrdrInv			OrderCommand = -1
	CostReq 		OrderCommand = 0
	CostSend 		OrderCommand = 1
	OrdrAssign 		OrderCommand = 2
	OrdrAdd 		OrderCommand = 3
	OrdrConf 		OrderCommand = 4
	OrdrDelete		OrderCommand = 5
	OrdrRetrans		OrderCommand = 6
)

type ElevCommand int
const (
  GoUp    			ElevCommand = 0
  GoDown    		ElevCommand = 1
  FloorReached  	ElevCommand = 2
  Finished  		ElevCommand = 3
)
