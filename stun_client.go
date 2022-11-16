package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/ahlixinjie/go-utils/collections/pair"
	"github.com/sirupsen/logrus"
)

type StunClient struct {
	conn net.Conn
}

func (s *StunClient) GetAddress() (IPPort *pair.Pair[string, string], err error) {
	data, id := packRequestSTUNData()
	_, err = s.conn.Write(data)
	if err != nil {
		logrus.Errorf("write STUN data error:%v", err)
		return
	}

	result := make([]byte, 1000)
	_, err = s.conn.Read(result)
	if err != nil {
		logrus.Errorf("read STUN data error:%v", err)
		return
	}

	replyID, replyPayload, err := unpackSTUNData(result)
	if err != nil {
		logrus.Errorf("unpackSTUNData error:%v", err)
		return
	}

	for i := 0; i < len(replyID); i++ {
		if id[i] != replyID[i] {
			err = errors.New("not same session")
			return
		}
	}
	IPPort, err = parseSTUNAttributes(replyPayload)
	if err != nil {
		logrus.Errorf("parseSTUNAttributes error:%v", err)
		return
	}
	return
}

/*
Comprehension-required range (0x0000-0x7FFF):

	0x0000: (Reserved)
	0x0001: MAPPED-ADDRESS
	0x0002: (Reserved; was RESPONSE-ADDRESS)
	0x0003: (Reserved; was CHANGE-ADDRESS)
	0x0004: (Reserved; was SOURCE-ADDRESS)
	0x0005: (Reserved; was CHANGED-ADDRESS)
	0x0006: USERNAME
	0x0007: (Reserved; was PASSWORD)
	0x0008: MESSAGE-INTEGRITY
	0x0009: ERROR-CODE
	0x000A: UNKNOWN-ATTRIBUTES
	0x000B: (Reserved; was REFLECTED-FROM)
	0x0014: REALM
	0x0015: NONCE
	0x0020: XOR-MAPPED-ADDRESS

Comprehension-optional range (0x8000-0xFFFF)

	0x8022: SOFTWARE
	0x8023: ALTERNATE-SERVER
	0x8028: FINGERPRINT
*/
func parseSTUNAttributes(data []byte) (result *pair.Pair[string, string], err error) {
	for len(data) >= 4 {
		attrType, attrLen := binary.BigEndian.Uint16(data[:2]), binary.BigEndian.Uint16(data[2:4])
		if len(data) < 4+int(attrLen) {
			err = errors.New("bad STUN attributes")
			return
		}
		val := data[4 : 4+int(attrLen)]
		if attrType == 0x0001 {
			/* MAPPED-ADDRESS
			    0                   1                   2                   3
			    0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
			   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			   |0 0 0 0 0 0 0 0|    Family     |           Port                |
			   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			   |                                                               |
			   |                 Address (32 bits or 128 bits)                 |
			   |                                                               |
			   +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			*/
			family := binary.BigEndian.Uint16(val[:2])
			if family == 0x0001 { //IPV4
				result = pair.NewPair(net.IP(val[4:8]).String(), fmt.Sprintf("%d", binary.BigEndian.Uint16(val[2:4])))
			}
			return
		} else if attrType == 0x0020 {
			//todo
			/* XOR-MAPPED-ADDRESS
			0                   1                   2                   3
			      0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
			     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			     |x x x x x x x x|    Family     |         X-Port                |
			     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			     |                X-Address (Variable)
			     +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			*/
			family := binary.BigEndian.Uint16(val[:2])
			if family == 0x0001 { //IPV4
				//xorPort := binary.BigEndian.Uint16(val[2:4])
				//xorIP := val[4:8]
			}
		}

		data = data[4+int(attrLen):]
	}
	return
}

func packRequestSTUNData() (data []byte, id []byte) {
	data = make([]byte, 20)
	binary.BigEndian.PutUint16(data[:2], 0x0001)      //first 2 bits zero + STUN Message Type BIND_REQUEST
	binary.BigEndian.PutUint16(data[2:4], 0)          //Message Length
	binary.BigEndian.PutUint32(data[4:8], 0x2112A442) //Magic Cookie
	//Transaction ID
	rand.Seed(time.Now().Unix())
	binary.BigEndian.PutUint32(data[8:12], rand.Uint32())
	binary.BigEndian.PutUint32(data[12:16], rand.Uint32())
	binary.BigEndian.PutUint32(data[16:20], rand.Uint32())

	return data, data[4:]
}

func unpackSTUNData(data []byte) (id []byte, payload []byte, err error) {
	if len(data) < 20 {
		err = errors.New("bad response")
		return
	}
	msgType, msgLength := binary.BigEndian.Uint16(data[:2]), binary.BigEndian.Uint16(data[2:4])

	if len(data) < 20+int(msgLength) {
		err = errors.New("bad response")
		return
	}

	if msgType != uint16(0x0101) { //first 2 bits zero + STUN Message Type BIND_RESPONSE
		err = errors.New("unknown stun msg type")
		return
	}
	return data[4:20], data[20 : 20+int(msgLength)], nil
}
