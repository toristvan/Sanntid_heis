package config

import (
	"fmt"
	"time"
)

const Ground_floor int = 0
const Num_floors int = 4
const Max_cost int = Num_floors
const Num_elevs int = 3

const Order_port int = 20005
const Backup_port int = 20070
const Peer_port int = 20024
const Offline_port int = 20003

var Local_ID int
var Current_floor int


func InitConfigData(id int){
	Local_ID   = id
	fmt.Println("Configuration data initialized")
}

type ElevStateType int
const (
	Idle		ElevStateType 	= 0
	GoingUp 					= 1
	GoingDown 	 				= 2
	AtFloor 		 			= 3
)

type MotorDirection int
const (
	MD_Up   	MotorDirection 	= 1
	MD_Down                		= -1
	MD_Stop                		= 0
)

type ButtonType int
const (
	BT_HallUp		ButtonType	= 0
	BT_HallDown            		= 1
	BT_Cab                 		= 2
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
	CostReq 					 = 0
	CostSend 					 = 1
	OrdrAssign 					 = 2
	OrdrAdd 					 = 3
	OrdrConf 					 = 4
	OrdrDelete					 = 5
	OrdrRetrans					 = 6
)

type ElevCommand int
const (
  GoUp 				ElevCommand = 0
  GoDown    					= 1
  FloorReached  				= 2
  Finished  					= 3
)
