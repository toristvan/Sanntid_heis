package config

import (
	"fmt"
	"time"
)

var LocalID int
var Num_elevs int
var Order_port int
var Backup_port	int

func InitConfigData(id int, num_of_elevs int){
	LocalID   = id
	Num_elevs = num_of_elevs
	Backup_port	= 20070 + LocalID
	Order_port = 20005
	fmt.Println("Configuration data initiated")
}

//const LocalID int = 0
//const Num_elevs int = 1
//var Num_elevs int

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
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
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
	ElevID int
	Cost int
	Cmd OrderCommand
	MasterID int
}

type OrderCommand int
const (
	CostReq 		OrderCommand = 0
	CostSend 		OrderCommand = 1
	OrdrAssign 		OrderCommand = 2
	OrdrAdd 		OrderCommand = 3
	OrdrConf 		OrderCommand = 4
	OrdrDelete		OrderCommand = 5
)

type ElevCommand int
const (
  //NewOrder  	ElevCommand = 0
  GoUp    			ElevCommand = 1
  GoDown    		ElevCommand = 2
  FloorReached  	ElevCommand = 3
  Finished  		ElevCommand = 4
)

//What kind of status?

/*
type Status int
const (
  Pending Status = 0
  Active  Status = 1
  Done    Status = 2
)
*/
