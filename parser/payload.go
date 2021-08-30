package parser

import (
	"encoding/binary"
	"fmt"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/adv"
)

// BLE payload including MSD parsing result
type Payload struct {
	// The ble/adv.Packet instense
	packet adv.Packet
	// The manufacturer specified data array
	msdata map[string]interface{}
}

// Parser entry
// Input payload ([]byte) and returns the Payload instense
func ParseBlePayload(bytes []byte) *Payload {
	pkt := adv.NewRawPacket(bytes)
	payload := Payload{*pkt, map[string]interface{}{}}
	payload.ibs() // call ibs parser
	return &payload
}

// Wrap to adv.Packet's ManufacturerData method
func (payload Payload) ManufacturerData() []byte {
	return payload.packet.ManufacturerData()
}

// Wrap to adv.Packet's ServiceData method
func (payload Payload) ServiceData() []ble.ServiceData {
	return payload.packet.ServiceData()
}

// Wrap to adv.Packet's LocalName method
func (payload Payload) LocalName() (string, bool) {
	// 0x08: shortened local name
	// 0x09: complete local name
	if b := payload.packet.Field(0x08); b != nil {
		return string(b), true
	}
	if b := payload.packet.Field(0x09); b != nil {
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

// Return humidity sensor reading (in %)
func (payload Payload) Humidity() (value int16, ok bool) {
	return payload.readingInt(fieldHumidity)
}

// Return range sensor reading (in %)
func (payload Payload) Range() (value int16, ok bool) {
	return payload.readingInt(fieldRange)
}

// Return GP sensor reading
func (payload Payload) GP() (value uint16, ok bool) {
	return payload.readingUint(fieldGP)
}

// Return sensor triggered counter
func (payload Payload) Counter() (value uint16, ok bool) {
	return payload.readingUint(fieldCounter)
}

// Return CO2 sensor reading (in ppm)
func (payload Payload) CO2() (value uint16, ok bool) {
	return payload.readingUint(fieldCO2)
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

// Return if external dim triggered
func (payload Payload) DinTriggered() (value bool, ok bool) {
	return payload.EventStat(evtDin)
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

// Stringer interface for Payload
func (payload Payload) String() string {
	var x []string
	if vendor, ok := payload.Vendor(); ok {
		x = append(x, vendor)
	}
	if model, ok := payload.ProductModel(); ok {
		x = append(x, model)
	}
	fieldList := []string{
		fieldBattery,
		fieldTemperature,
		fieldTempExt,
		fieldHumidity,
		fieldRange,
		fieldCounter,
		fieldCO2,
		fieldGP,
		fieldUserData,
		fieldAccel,
		fieldAccels,
		evtButton,
		evtHall,
		evtFall,
		evtMoving,
		evtIR,
		evtPIR,
		evtDin,
	}
	for _, f := range fieldList {
		if v, ok := payload.msdata[f]; ok {
			x = append(x, fmt.Sprintf("%v: %v", f, v))
		}
	}
	return fmt.Sprintf("%v\n", x)
}
