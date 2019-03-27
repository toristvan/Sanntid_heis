package peers

import (
	"../conn"
	"fmt"
	"net"
	"sort"
	"time"
)

type PeerUpdate struct {
	Peers []string
	New   string
	Lost  []string
}

const interval = 1000 * time.Millisecond
const timeout = 50 * time.Millisecond


//What does this currently do?

func Transmitter(port int, id string, offline_check chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	//enable := true
	for {
			//enable = <-offline_check
			transmitTicker := time.NewTicker(interval)
			select {
			//case <-time.After(interval):
			case <- transmitTicker.C:
				//if enable {
					_, err := conn.WriteTo([]byte(id), addr)
					if err != nil {
							fmt.Println(err)
					}
				//}
			}
		}
}



//function to check if node can connect to router
//Logic for not writing to channel continously can be implemented
func CheckOffline(port int, offline_chan chan<- bool){
	var offline bool = false
	for {
		addr,_ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
		_, err := net.DialUDP("udp4", nil, addr)
		if err != nil{
			if !offline{
				offline = true
				offline_chan <- offline
			}
		} else {
			if offline{
				offline = false
				offline_chan <- offline
			}
		}
		time.Sleep(1*time.Second)
	}
}

func Receiver(port int, peerUpdateCh chan<- PeerUpdate) {

	var buf [1024]byte
	var p PeerUpdate
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id := string(buf[:n])

		// Adding new connection
		p.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			sort.Strings(p.Peers)
			sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}
