package queue

type OrderCommand int 

const (
	CostReq 	OrderCommand = 0
	CostSend 	OrderCommand = 1
	OrdrAssign 	OrderCommand = 2
	OrdrAdd 	OrderCommand = 3
	OrdrConf 	OrderCommand = 4
)

type OrderStruct struct
{
	Button elevio.ButtonType
	Floor int
	timestamp time.Time
	ElevID int
	Cost int
	Cmd OrderCommand

}