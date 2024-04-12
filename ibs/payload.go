package ibs

import (
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/adv"
)

// BLE payload including MSD parsing result
type Payload struct {
	// The ble/adv.Packet instense
	Packet adv.Packet
	// The manufacturer specified data array
	msdata map[string]interface{}
}

// Parser entry
// Input payload ([]byte) and returns the Payload instense
func Parse(bytes []byte) *Payload {
	pkt := adv.NewRawPacket(bytes)
	payload := Payload{*pkt, map[string]interface{}{}}
	payload.ibs() // call ibs parser
	return &payload
}

// Wrap to adv.Packet's ManufacturerData method
func (payload Payload) ManufacturerData() []byte {
	return payload.Packet.ManufacturerData()
}

// Wrap to adv.Packet's ServiceData method
func (payload Payload) ServiceData() []ble.ServiceData {
	return payload.Packet.ServiceData()
}

// Wrap to adv.Packet's LocalName method
func (payload Payload) LocalName() (string, bool) {
	// 0x08: shortened local name
	// 0x09: complete local name
	if b := payload.Packet.Field(0x08); b != nil {
		return string(b), true
	}
	if b := payload.Packet.Field(0x09); b != nil {
		return string(b), true
	}
	return "", false
}

// Returns vendor code (mfg) got from manufacturer data
func (payload Payload) VendorCode() (code uint16, ok bool) {
	if msd := payload.ManufacturerData(); msd != nil {
		return binary.LittleEndian.Uint16(msd[:2]), true
	}
	return 0, false
}

// Returns vendor name got from manufacturer data
func (payload Payload) Vendor() (name string, ok bool) {
	if code, ok := payload.VendorCode(); ok {
		// check parsed mas first,
		// it may contains heck for Ingics Beacon
		if name, ok := payload.msdata["vendor"]; ok {
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
func (payload Payload) ProductModel() (name string, ok bool) {
	if mfg, ok := payload.VendorCode(); ok {
		msd := payload.ManufacturerData()
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
		if model, ok := payload.msdata["model"]; ok {
			return model.(string), true
		}
	}
	return "", false
}

// Helper function for query sensor reading
func (payload Payload) readingInt(sensor string) (value int16, ok bool) {
	if value, ok := payload.msdata[sensor]; ok {
		return value.(int16), true
	}
	return 0, false
}

// Helper function for query sensor reading
func (payload Payload) readingUint(sensor string) (value uint16, ok bool) {
	if value, ok := payload.msdata[sensor]; ok {
		return value.(uint16), true
	}
	return 0, false
}

// Helper function for query sensor reading
func (payload Payload) readingFloat(sensor string) (value float32, ok bool) {
	if value, ok := payload.msdata[sensor]; ok {
		return value.(float32), true
	}
	return 0, false
}

// Return battery voltage (in V)
func (payload Payload) BatteryVoltage() (value float32, ok bool) {
	return payload.readingFloat(fieldBattery)
}

// Return temperature sensor reading (in C)
func (payload Payload) Temperature() (value float32, ok bool) {
	return payload.readingFloat(fieldTemperature)
}

// Return external temperature sensor reading (in C)
func (payload Payload) TemperatureExt() (value float32, ok bool) {
	return payload.readingFloat(fieldTempExt)
}

// Return external temperature sensor reading (in C)
func (payload Payload) TemperatureEnv() (value float32, ok bool) {
	return payload.readingFloat(fieldTempEnv)
}

// Return humidity sensor reading (in %)
func (payload Payload) Humidity() (value float32, ok bool) {
	return payload.readingFloat(fieldHumidity)
}

// Return range sensor reading (in %)
func (payload Payload) Range() (value int, ok bool) {
	if v, ok := payload.readingInt(fieldRange); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// Return GP sensor reading
func (payload Payload) GP() (value float32, ok bool) {
	return payload.readingFloat(fieldGP)
}

// Return sensor triggered counter
func (payload Payload) Counter() (value int, ok bool) {
	if v, ok := payload.readingUint(fieldCounter); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// Return CO2 sensor reading (in ppm)
func (payload Payload) CO2() (value int, ok bool) {
	if v, ok := payload.readingUint(fieldCO2); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// Return voltage sensor reading (in mV)
func (payload Payload) Voltage() (value int, ok bool) {
	if v, ok := payload.readingInt(fieldVoltage); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// Return current sensor reading (in µA)
func (payload Payload) Current() (value uint, ok bool) {
	if v, ok := payload.readingUint(fieldCurrent); ok {
		return uint(v), true
	} else {
		return 0, false
	}
}

// Return event stat
func (payload Payload) EventStat(evt string) (value bool, ok bool) {
	if value, ok := payload.msdata[evt]; ok {
		return value.(bool), true
	}
	return false, false
}

// Return if button pressed
func (payload Payload) ButtonPressed() (value bool, ok bool) {
	return payload.EventStat(evtButton)
}

// Return if moving detected
func (payload Payload) Moving() (value bool, ok bool) {
	return payload.EventStat(evtMoving)
}

// Return if hall sensor detected
func (payload Payload) HallDetected() (value bool, ok bool) {
	return payload.EventStat(evtHall)
}

// Return if falling detected
func (payload Payload) Falling() (value bool, ok bool) {
	return payload.EventStat(evtFall)
}

// Return if PIR sensor detected
func (payload Payload) PIRDetected() (value bool, ok bool) {
	return payload.EventStat(evtPIR)
}

// Return if IR sensor detected
func (payload Payload) IRDetected() (value bool, ok bool) {
	return payload.EventStat(evtIR)
}

// Return if IR sensor detected
func (payload Payload) Detected() (value bool, ok bool) {
	return payload.EventStat(evtDetect)
}

// Return if external din triggered
func (payload Payload) DinTriggered() (value bool, ok bool) {
	return payload.EventStat(evtDin)
}

// Return if external din2 triggered
func (payload Payload) Din2Triggered() (value bool, ok bool) {
	return payload.EventStat(evtDin2)
}

// Return accel readings
func (payload Payload) Accel() (reading AccelReading, ok bool) {
	if value, ok := payload.msdata[fieldAccel]; ok {
		return value.(AccelReading), true
	}
	return AccelReading{0, 0, 0}, false
}

// Return accels readings
func (payload Payload) Accels() (reading []AccelReading, ok bool) {
	if value, ok := payload.msdata[fieldAccels]; ok {
		return value.([]AccelReading), true
	}
	return []AccelReading{}, false
}

// return lux reading
func (payload Payload) Lux() (reading int, ok bool) {
	if v, ok := payload.readingInt(fieldLux); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// return value field
func (payload Payload) Value() (reading uint, ok bool) {
	if v, ok := payload.readingUint(fieldValue); ok {
		return uint(v), true
	} else {
		return 0, false
	}
}

// return user data
func (payload Payload) UserData() (reading int, ok bool) {
	if v, ok := payload.readingInt(fieldUserData); ok {
		return int(v), true
	} else {
		return 0, false
	}
}

// Stringer interface for Payload
func (payload Payload) String() string {
	var x []string
	for k, v := range payload.msdata {
		x = append(x, fmt.Sprintf("%v: %v", k, v))
	}
	return fmt.Sprintf("{ %v }", strings.Join(x, ", "))
}
