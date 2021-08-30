package parser

import (
	"encoding/binary"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/adv"
)

// Packet is an implementation of adv.AdvPacket for crafting or parsing manufacturer data
type Packet struct {
	adv adv.Packet
	msd map[string]interface{}
}

// Warp of adv.NewRawPacket() to return Packet instense
func ParseBlePayload(payload []byte) *Packet {
	adv := adv.NewRawPacket(payload)
	packet := Packet{*adv, map[string]interface{}{}}
	packet.ibs() // call ibs parser
	return &packet
}

// Wrap to adv.Packet's ManufacturerData method
func (pkt Packet) ManufacturerData() []byte {
	return pkt.adv.ManufacturerData()
}

// Wrap to adv.Packet's ServiceData method
func (pkt Packet) ServiceData() []ble.ServiceData {
	return pkt.adv.ServiceData()
}

// Wrap to adv.Packet's LocalName method
func (pkt Packet) LocalName() (string, bool) {
	// 0x08: shortened local name
	// 0x09: complete local name
	if b := pkt.adv.Field(0x08); b != nil {
		return string(b), true
	}
	if b := pkt.adv.Field(0x09); b != nil {
		return string(b), true
	}
	return "", false
}

// Returns vendor code (mfg) got from manufacturer data
func (pkt Packet) VendorCode() (code uint16, ok bool) {
	if msd := pkt.ManufacturerData(); msd != nil {
		return binary.LittleEndian.Uint16(msd[:2]), true
	}
	return 0, false
}

// Returns vendor name got from manufacturer data
func (pkt Packet) Vendor() (name string, ok bool) {
	if code, ok := pkt.VendorCode(); ok {
		// check parsed mas first,
		// it may contains heck for Ingics Beacon
		if name, ok := pkt.msd["vendor"]; ok {
			return name.(string), true
		}
		// query known vendor code if not Ingics beacon
		if name, ok := knownVendorCode[code]; ok {
			return name, true
		}
		// unknown mfg value, using the HEX number
		return fmt.Sprintf("0x%04X", code), true
	}
	return "", false
}

// Returns product type (model) guess from manufacturer data
func (pkt Packet) ProductModel() (name string, ok bool) {
	if mfg, ok := pkt.VendorCode(); ok {
		msd := pkt.ManufacturerData()
		if mfg == 0x0006 { // Microsoft
			typ := uint8(msd[3]) & 0x3F
			if name, ok := microsoftProductType[typ]; ok {
				return name, true
			}
			return "", false
		} else if mfg == 0x004C { // Apple
			if msd[2] == 0x02 && len(msd) == 25 {
				return "iBeacon", true
			}
			return "", false
		}
		if model, ok := pkt.msd["model"]; ok {
			return model.(string), true
		}
	}
	return "", false
}

// Helper function for query sensor reading
func (pkt Packet) readingInt(sensor string) (value int16, ok bool) {
	if value, ok := pkt.msd[sensor]; ok {
		return value.(int16), true
	}
	return 0, false
}

// Helper function for query sensor reading
func (pkt Packet) readingUint(sensor string) (value uint16, ok bool) {
	if value, ok := pkt.msd[sensor]; ok {
		return value.(uint16), true
	}
	return 0, false
}

// Helper function for query sensor reading
func (pkt Packet) readingFloat(sensor string) (value float32, ok bool) {
	if value, ok := pkt.msd[sensor]; ok {
		return value.(float32), true
	}
	return 0, false
}

// Return battery voltage (in V)
func (pkt Packet) BatteryVoltage() (value float32, ok bool) {
	return pkt.readingFloat(fieldBattery)
}

// Return temperature sensor reading (in C)
func (pkt Packet) Temperature() (value float32, ok bool) {
	return pkt.readingFloat(fieldTemperature)
}

// Return external temperature sensor reading (in C)
func (pkt Packet) TemperatureExt() (value float32, ok bool) {
	return pkt.readingFloat(fieldTempExt)
}

// Return humidity sensor reading (in %)
func (pkt Packet) Humidity() (value int16, ok bool) {
	return pkt.readingInt(fieldHumidity)
}

// Return range sensor reading (in %)
func (pkt Packet) Range() (value int16, ok bool) {
	return pkt.readingInt(fieldRange)
}

// Return GP sensor reading
func (pkt Packet) GP() (value uint16, ok bool) {
	return pkt.readingUint(fieldGP)
}

// Return sensor triggered counter
func (pkt Packet) Counter() (value uint16, ok bool) {
	return pkt.readingUint(fieldCounter)
}

// Return CO2 sensor reading (in ppm)
func (pkt Packet) CO2() (value uint16, ok bool) {
	return pkt.readingUint(fieldCO2)
}

// Return event stat
func (pkt Packet) EventStat(evt string) (value bool, ok bool) {
	if value, ok := pkt.msd[evt]; ok {
		return value.(bool), true
	}
	return false, false
}

// Return if button pressed
func (pkt Packet) ButtonPressed() (value bool, ok bool) {
	return pkt.EventStat(evtButton)
}

// Return if moving detected
func (pkt Packet) Moving() (value bool, ok bool) {
	return pkt.EventStat(evtMoving)
}

// Return if hall sensor detected
func (pkt Packet) HallDetected() (value bool, ok bool) {
	return pkt.EventStat(evtHall)
}

// Return if falling detected
func (pkt Packet) Falling() (value bool, ok bool) {
	return pkt.EventStat(evtFall)
}

// Return if PIR sensor detected
func (pkt Packet) PIRDetected() (value bool, ok bool) {
	return pkt.EventStat(evtPIR)
}

// Return if IR sensor detected
func (pkt Packet) IRDetected() (value bool, ok bool) {
	return pkt.EventStat(evtIR)
}

// Return if external dim triggered
func (pkt Packet) DinTriggered() (value bool, ok bool) {
	return pkt.EventStat(evtDin)
}

// Return accel readings
func (pkt Packet) Accel() (reading AccelReading, ok bool) {
	if value, ok := pkt.msd[fieldAccel]; ok {
		return value.(AccelReading), true
	}
	return AccelReading{0, 0, 0}, false
}

// Return accels readings
func (pkt Packet) Accels() (reading []AccelReading, ok bool) {
	if value, ok := pkt.msd[fieldAccels]; ok {
		return value.([]AccelReading), true
	}
	return []AccelReading{}, false
}
