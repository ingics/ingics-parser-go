package igs

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

type Message struct{ s string }

const msgPattern = `^\$(.+),([0-9a-fA-F]{12}),([0-9a-fA-F]{12}),(-?\d+),(.*)$`

// Message constructor, input message got from iGS device
func Parse(s string) *Message {
	// clone string and create message
	b := make([]byte, len(s))
	copy(b, s)
	m := &Message{strings.TrimSpace(*(*string)(unsafe.Pointer(&b)))}
	// validation
	if m.validate() {
		return m
	}
	return nil
}

// Stringer of Messsage
func (m Message) String() string {
	return m.s
}

func (m Message) validate() bool {
	p := regexp.MustCompile(msgPattern)
	return p.MatchString(m.s)
}

func (m Message) fields() []string {
	p := regexp.MustCompile(msgPattern)
	return p.FindStringSubmatch(m.s)[1:]
}

// Message type
// GPRP: BLE4.2 General Purpose Report
// RSPR: BLE4.2 Scan Response Report
// LRAD: BLE 5 Long Range ADV
// LRSR: BLE 5 Long Range Scan Response
// 1MAD: BLE 5 1M ADV
// 1MSR: BLE 5 1M Scan Response
func (m Message) MsgType() string {
	return m.fields()[0]
}

// Beacon (Tag) BLE mac address
func (m Message) Beacon() string {
	return m.fields()[1]
}

// Gateway (IGSXX) mac address
func (m Message) Gateway() string {
	return m.fields()[2]
}

// RSSI value
func (m Message) RSSI() int {
	if rssi, err := strconv.Atoi(m.fields()[3]); err == nil {
		return rssi
	} else {
		return -127
	}
}

// BLE payload in HEX string
func (m Message) Payload() string {
	return strings.Split(m.fields()[4], ",")[0]
}

// Timestamp append in message, maybe nil
func (m Message) Timestamp() *time.Time {
	if x := strings.Split(m.fields()[4], ","); len(x) > 1 {
		sec := 0
		msec := 0
		if y := strings.Split(x[1], "."); len(y) > 1 {
			sec, _ = strconv.Atoi(y[0])
			msec, _ = strconv.Atoi(y[1])
		} else {
			sec, _ = strconv.Atoi(y[0])
		}
		t := time.Unix(int64(sec), int64(msec*1000))
		return &t
	}
	return nil
}
