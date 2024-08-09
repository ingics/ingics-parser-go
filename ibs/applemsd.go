package ibs

import (
	"encoding/binary"
)

func (pkt Payload) apple() bool {
	// Apple iBeacon
	msd := pkt.ManufacturerData()
	if len(msd) == 25 && msd[0] == 0x4C && msd[1] == 0x00 && msd[2] == 0x02 {
		pkt.msdata["uuid"] = msd[4:20]
		pkt.msdata["major"] = binary.BigEndian.Uint16(msd[20:22])
		pkt.msdata["minor"] = binary.BigEndian.Uint16(msd[22:24])
		pkt.msdata["txpower"] = int16(int8(msd[24]))
		return true
	}
	return false
}
