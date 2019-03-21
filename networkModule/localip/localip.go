package localip

import (
	."fmt"
	"net"
	"strings"
	"flag"
	"os"
)

var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func SetPID() string {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := LocalIP()
		if err != nil {
			Println(err)
			localIP = "DISCONNECTED"
		}
		id = Sprintf("%s-%d",localIP, os.Getpid())
	}
	return id
}
