package peers

import (
	"../conn"
	"fmt"
	"net"
	//"sort"
	"time"
	"strconv"
)

type PeerUpdate struct {
	Peers []int
	New   int
	Lost  []int
}

var ActivePeers PeerUpdate
const interval = 200* time.Millisecond
const timeout = 1000 * time.Millisecond


//What does this currently do?

func Transmitter(port int, id int, transmit_enable_chan chan bool) {

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {

		case enable = <- transmit_enable_chan:
		case <- time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(strconv.Itoa(id)), addr)
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
	lastSeen := make(map[int]time.Time)

	conn := conn.DialBroadcastUDP(port)

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buf[0:])

		id, _ := strconv.Atoi(string(buf[:n]))

		// Adding new connection
		p.New = -1
		if id >-1 {
			if _, idExists := lastSeen[id]; !idExists {
				p.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		p.Lost = make([]int, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				p.Lost = append(p.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			p.Peers = make([]int, 0, len(lastSeen))

			for k, _ := range lastSeen {
				p.Peers = append(p.Peers, k)
			}

			//sort.Strings(p.Peers)
			//sort.Strings(p.Lost)
			peerUpdateCh <- p
		}
	}
}


func CheckForPeers(peers_update_chan <-chan PeerUpdate){
	for{
		ActivePeers = <-peers_update_chan
		fmt.Printf("Peer update:\n")
		fmt.Printf("  Peers num:    %d\n", len(ActivePeers.Peers))
		fmt.Printf("  Peers:    %d\n", ActivePeers.Peers)
		fmt.Printf("  New:      %d\n", ActivePeers.New)
		fmt.Printf("  Lost:     %d\n", ActivePeers.Lost)
	}	

}
