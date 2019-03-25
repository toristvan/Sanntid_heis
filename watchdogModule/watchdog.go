package watchdog

import(
  "../queueModule"
  "../networkModule/bcast"
  ."fmt"
  "time"
  )

/*
type WatchdogTimer struct {
  startTime     time.Time
  timeNow       time.Duration
  timeoutLength time.Duration
}

func NewWatchdog() WatchdogTimer {
	wtd := new(WatchdogTimer)
	return *wtd
}

func (wtd *WatchdogTimer) initWatchdog(timeOutInterval time.Duration) {
  wtd.startTime     = time.Now()
  wtd.timeoutLength = timeOutInterval
}

func (wtd *WatchdogTimer) resetWatchdog() {
  wtd.startTime = time.Now()
}


func (wtd *WatchdogTimer) timeOut() bool{
  return time.Since(wtd.startTime)>wtd.timeoutLength
}
*/
func timeout (order *orderStruct) bool{
  return time.Since(order.startTime)>length;
}

func Watchdog(retransmit chan<- config.OrderStruct, numElevs int, queueSize int){
  //a := NewWatchdog()
  //a.initWatchdog(10 * time.Second)

  for{
    time.Sleep(5 * time.Second)
    //a.updateWatchdog()
    //a.resetWatchdog()
    for j := 0; j< numElevs; j++{
      for i := 0; i< queueSize; i++{
        if timeout(queue.OrderQueue[j][i]) {
          retransmit <- queue.OrderQueue[j][i]
        }
      }
    }

  }
}
