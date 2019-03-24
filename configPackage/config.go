package config

import (
	"time"
)


type ElevStateType int
const (
	Idle 				ElevStateType = 0
	GoingUp 			ElevStateType = 1
	GoingDown 			ElevStateType = 2
	AtFloor 			ElevStateType = 3
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
}

type OrderCommand int 

const (
	CostReq 	OrderCommand = 0
	CostSend 	OrderCommand = 1
	OrdrAssign 	OrderCommand = 2
	OrdrAdd 	OrderCommand = 3
	OrdrConf 	OrderCommand = 4
)

type ElevCommand int
const (
  //NewOrder  	ElevCommand = 0
  GoUp    		ElevCommand = 1
  GoDown    	ElevCommand = 2
  FloorReached  ElevCommand = 3
  Finished  	ElevCommand = 4
  Wait          ElevCommand = 5
)

//What kind of status?
type Status int
const (
  Pending Status = 0
  Active  Status = 1
  Done    Status = 2
)