package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"testing"
)

func Test_packRequestStunData(t *testing.T) {
	fmt.Println(packRequestSTUNData())
}

func Test_stunConnection(t *testing.T) {
	conn, err := net.Dial("tcp4", "stun.voip.blackberry.com:3478")
	if err != nil {
		t.Errorf(err.Error())
	}
	defer conn.Close()

	client := StunClient{conn}
	IPPort, err := client.GetAddress()
	if err != nil {
		t.Error(err)
		return
	}
	logrus.Info(IPPort.First, IPPort.Second)
}
