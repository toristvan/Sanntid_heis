package main

import(
  ."fmt"
  "time"
)

type Watchdogtimer struct {
  startTime     time.Time
  timeNow       time.Duration
  timeoutLength time.Duration
}

func NewWatchdog() Watchdogtimer {
	wtd := new(Watchdogtimer)
	return *wtd
}

func (wtd *Watchdogtimer) initWatchdog(timeOutInterval time.Duration) {
  wtd.startTime     = time.Now()
  wtd.timeoutLength = timeOutInterval
}

func (wtd *Watchdogtimer) resetWatchdog() {
  wtd.startTime = time.Now()
}

func (wtd *Watchdogtimer) updateWatchdog() {
  wtd.timeNow = time.Since(wtd.startTime)
  Println("time now:", wtd.timeNow)
  if wtd.timeNow > wtd.timeoutLength {
    Println("timeout")
  }
}

func main(){
  a := NewWatchdog()
  a.initWatchdog(10 * time.Second)

  for{
    time.Sleep(1 * time.Second)
    a.updateWatchdog()
    a.resetWatchdog()
  }
}
