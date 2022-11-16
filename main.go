package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	_ "github.com/ahlixinjie/mongoose/log"
	"github.com/libp2p/go-reuseport"
	"github.com/sirupsen/logrus"
)

func main() {
	//test()
	contact()
}
func contact() {
	newConn, err := reuseport.Dial("tcp4", "0.0.0.0:10005", "www.baidu.com:80")
	if err != nil {
		panic(err)
	}
	defer newConn.Close()

	logrus.Infof("local IP addrs:%s", newConn.LocalAddr().String())
	var httpData strings.Builder
	httpData.WriteString("HEAD / HTTP/1.1\r\n")
	httpData.WriteString("HOST: www.baidu.com\r\n")
	httpData.WriteString("Connection: keep-alive\r\n")
	httpData.WriteString("\r\n")

	for i := 0; i < 1; i++ {
		time.Sleep(time.Second * 5)
		now := time.Now()
		if err = newConn.SetWriteDeadline(now.Add(time.Second * 2)); err != nil {
			panic(err)
		}
		if _, err = io.WriteString(newConn, httpData.String()); err != nil {
			panic(err)
		}

		now = time.Now()
		if err = newConn.SetReadDeadline(now.Add(time.Second * 3)); err != nil {
			panic(err)
		}
		var result = make([]byte, 1024)
		_, err = newConn.Read(result)
		fmt.Print(string(result))

		time.Sleep(time.Second * 5)
	}

}
