package ibs

import (
	"encoding/binary"
)

const IngicsVendorCode = 0x082C

type AccelReading struct {
	x float32
	y float32
	z float32
}

const (
	fieldModel       = "model"
	fieldVendor      = "vendor"
	fieldBattery     = "battery"
	fieldTemperature = "temperature"
	fieldHumidity    = "humidity"
	fieldTempExt     = "temperatureExt"
	fieldRange       = "range"
	fieldGP          = "gp"
	fieldCounter     = "counter"
	fieldCO2         = "co2"
	fieldAccel       = "accel"
	fieldAccels      = "accels"
	fieldUserData    = "userdata"
	fieldEvents      = "events"
	fieldSubtype     = "subtype"
	fieldReserved    = "reserved"
	fieldReserved2   = "reserved2"
	fieldBattAct     = "battact"
	fieldRsEvents    = "rsEvents"
)

const (
	evtButton = "button"
	evtMoving = "moving"
	evtHall   = "hall"
	evtFall   = "fall"
	evtPIR    = "pir"
	evtIR     = "ir"
	evtDin    = "din"
	bitButton = 0
	bitMoving = 1
	bitHall   = 2
	bitFall   = 3
	bitPIR    = 4
	bitIR     = 5
	bitDin    = 6
)

var evtMasks = map[string]uint8{
	evtButton: uint8(1 << bitButton),
	evtMoving: uint8(1 << bitMoving),
	evtHall:   uint8(1 << bitHall),
	evtFall:   uint8(1 << bitFall),
	evtPIR:    uint8(1 << bitPIR),
	evtIR:     uint8(1 << bitIR),
	evtDin:    uint8(1 << bitDin),
}

func (pkt Payload) ibs() bool {
	if mfg, ok := pkt.VendorCode(); ok {
		msd := pkt.packet.ManufacturerData()
		code := binary.LittleEndian.Uint16(msd[2:4])
		if mfg == 0x59 && code == 0xBC80 {
			// iBS01(H/G/T)
			return pkt.ibs01()
		} else if code == 0xBC81 {
			// iBS01RG
			var rgPayloadDef = payloadDef{
				pkt.rgNaming,
				[]string{fieldBattAct, fieldAccels},
				[]string{},
			}
			return pkt.parsePayload(rgPayloadDef)
		} else if code == 0xBC82 {
			// iBS02 for RS
			return pkt.parsePayloadBySubtype(rsPayloadDefs)
		} else if mfg == 0x0D && code == 0xBC83 {
			// iBS02/iBS03/iBS04 common payload
			return pkt.parsePayloadBySubtype(ibsCommonPayloadDefs)
		} else if mfg == 0x0D && code == 0x0BC85 {
			// iBS03GP
			var gpPayloadDef = payloadDef{
				"iBS03GP",
				[]string{fieldBattAct, fieldAccels, fieldGP},
				[]string{},
			}
			return pkt.parsePayload(gpPayloadDef)
		} else if mfg == 0x082C && code == 0xBC83 {
			// iBS05/iBS06
			return pkt.parsePayloadBySubtype(ibsCommonPayloadDefs)
		}
	}
	return false
}

func (pkt Payload) ibs01() bool {
	if mfg, ok := pkt.VendorCode(); ok {
		msd := pkt.ManufacturerData()
		typ := binary.LittleEndian.Uint16(msd[2:4])
		if mfg == 0x59 && typ == 0xBC80 {
			subtype := uint8(msd[13])
			if subtype == 0xff || subtype == 0x00 {
				// old firmware without subtype
				pkt.msdata[fieldVendor] = knownVendorCode[IngicsVendorCode]
				pkt.handleFloatField(fieldBattery, 4)
				if len(msd) > 7 && msd[7] != 0xFF && msd[8] != 0xFF {
					// has temperature value, should be iBS01T
					pkt.msdata[fieldModel] = "iBS01T"
					pkt.handleFloatField(fieldTemperature, 7)
					pkt.handleIntField(fieldHumidity, 9)
				} else {
					// others, cannot detect the 'read' model from payload
					// list all possible sensor fields
					flags := uint8(msd[6])
					pkt.msdata[fieldModel] = "iBS01"
					pkt.msdata[evtButton] = flags&evtMasks[evtButton] != 0
					pkt.msdata[evtMoving] = flags&evtMasks[evtMoving] != 0
					pkt.msdata[evtHall] = flags&evtMasks[evtHall] != 0
					pkt.msdata[evtFall] = flags&evtMasks[evtFall] != 0
				}
				return true
			} else {
				return pkt.parsePayloadBySubtype(ibs01PayloadDefs)
			}
		}
	}
	return false
}

func (pkt Payload) handleIntField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xFFFF && unsignedValue != 0xAAAA {
		pkt.msdata[name] = int16(unsignedValue)
	}
	return index + 2
}

func (pkt Payload) handleUintField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xFFFF && unsignedValue != 0xAAAA {
		pkt.msdata[name] = unsignedValue
	}
	return index + 2
}

func (pkt Payload) handleFloatField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xFFFF && unsignedValue != 0xAAAA {
		pkt.msdata[name] = float32(int16(unsignedValue)) / 100.0
	}
	return index + 2
}

func (pkt Payload) handleAccelField(name string, index int) int {
	msd := pkt.ManufacturerData()
	pkt.msdata[name] = AccelReading{
		float32(int16(binary.LittleEndian.Uint16(msd[index:index+2]))) * 4 / 100,
		float32(int16(binary.LittleEndian.Uint16(msd[index+2:index+4]))) * 4 / 100,
		float32(int16(binary.LittleEndian.Uint16(msd[index+4:index+6]))) * 4 / 100,
	}
	return index + 6
}

func (pkt Payload) handleAccelsField(name string, index int) int {
	msd := pkt.ManufacturerData()
	pkt.msdata[name] = []AccelReading{
		{
			float32(int16(binary.LittleEndian.Uint16(msd[index:index+2]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+2:index+4]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+4:index+6]))) * 4 / 100,
		},
		{
			float32(int16(binary.LittleEndian.Uint16(msd[index+6:index+8]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+8:index+10]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+10:index+12]))) * 4 / 100,
		},
		{
			float32(int16(binary.LittleEndian.Uint16(msd[index+12:index+14]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+14:index+16]))) * 4 / 100,
			float32(int16(binary.LittleEndian.Uint16(msd[index+16:index+18]))) * 4 / 100,
		},
	}
	return index + 18
}

func (pkt Payload) handleByteField(name string, index int) int {
	msd := pkt.ManufacturerData()
	pkt.msdata[name] = uint8(msd[index])
	return index + 1
}

func (pkt Payload) handleReservedField(name string, index int) int {
	return index + 1
}

func (pkt Payload) handleReserved2Field(name string, index int) int {
	return index + 2
}

func (pkt Payload) handleGpField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	pkt.msdata[name] = float32(int16(unsignedValue)) / 50.0
	return index + 2
}

// special handler for RG models, two bytes present battery + events
// if will fill value of battery, event value & events fields
func (pkt Payload) handleBattActField(name string, index int) int {
	msd := pkt.ManufacturerData()
	value := binary.LittleEndian.Uint16(msd[index : index+2])
	pkt.msdata[fieldBattery] = float32(int32(value&0x00FFF)) / 100.0
	events := uint8(value & 0xF000 >> 12)
	pkt.msdata[fieldEvents] = events
	pkt.msdata[evtButton] = (events & 0x02) != 0
	pkt.msdata[evtMoving] = (events & 0x01) != 0
	return index + 2
}

// special handler for RS events field
func (pkt Payload) handleRsEventsField(name string, index int) int {
	msd := pkt.ManufacturerData()
	value := uint8(msd[index])
	pkt.msdata[fieldEvents] = value
	pkt.msdata[evtDin] = (value & 0x04) != 0
	return 1
}

// special handler for RG models,
// returns model name determine by mfg code and type
func (pkt Payload) rgNaming() string {
	if mfg, ok := pkt.VendorCode(); ok {
		msd := pkt.ManufacturerData()
		typ := binary.LittleEndian.Uint16(msd[2:4])
		if mfg == 0x59 {
			return "iBS01RG"
		} else if typ == 0xBC81 {
			return "iBS03RG"
		}
	}
	return "iBSXXRG"
}

type payloadDef struct {
	model  interface{}
	fields []string
	events []string
}

var ibs01PayloadDefs = map[byte]payloadDef{
	0x03: {
		"iBS01",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtButton},
	},
	0x04: {
		"iBS01H",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtButton, evtHall},
	},
	0x05: {
		"iBS01T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldHumidity, fieldReserved2},
		[]string{evtButton},
	},
	0x06: {
		"iBS01G",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtButton, evtMoving, evtFall},
	},
	0x07: {
		"iBS01T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldReserved2, fieldReserved2},
		[]string{evtButton},
	},
}

var rsPayloadDefs = map[byte]payloadDef{
	0x01: {
		"iBS02PIR-RS",
		[]string{fieldBattery, fieldRsEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
	0x02: {
		"iBS02IR-RS",
		[]string{fieldBattery, fieldRsEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
	0x04: {
		"iBS02HM",
		[]string{fieldBattery, fieldRsEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
}

var ibsCommonPayloadDefs = map[byte]payloadDef{
	0x01: {
		"iBS02PIR2",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtPIR},
	},
	0x02: {
		"iBS02IR2",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtIR},
	},
	0x04: {
		"iBS02M2",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtDin},
	},
	0x10: {
		"iBS03",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton, evtHall},
	},
	0x12: {
		"iBS03P",
		[]string{fieldBattery, fieldReserved, fieldTemperature, fieldTempExt, fieldUserData},
		[]string{},
	},
	0x13: {
		"iBS03R",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldRange, fieldUserData},
		[]string{},
	},
	0x14: {
		"iBS03T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldHumidity, fieldUserData},
		[]string{evtButton},
	},
	0x15: {
		"iBS03T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x16: {
		"iBS03G",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton, evtMoving, evtFall},
	},
	0x17: {
		"iBS03TP",
		[]string{fieldBattery, fieldReserved, fieldTemperature, fieldTempExt, fieldUserData},
		[]string{},
	},
	0x18: {
		"iBS04i",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x19: {
		"iBS04",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x20: {
		"iRS02",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldReserved2, fieldUserData},
		[]string{evtHall},
	},
	0x21: {
		"iRS02TP",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldTempExt, fieldUserData},
		[]string{evtHall},
	},
	0x22: {
		"iRS02RG",
		[]string{fieldBattery, fieldEvents, fieldAccel},
		[]string{evtHall},
	},
	0x30: {
		"iBS05",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x31: {
		"iBS05H",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton, evtHall},
	},
	0x32: {
		"iBS05T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x33: {
		"iBS05G",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton, evtMoving},
	},
	0x34: {
		"iBS05CO2",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCO2, fieldUserData},
		[]string{evtButton},
	},
	0x40: {
		"iBS06",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
}

// Parse the payload follow the input definition
func (pkt Payload) parsePayload(def payloadDef) bool {
	// define field handlers
	var fieldDefs = map[string]func(string, int) int{
		fieldBattery:     pkt.handleFloatField,
		fieldTemperature: pkt.handleFloatField,
		fieldTempExt:     pkt.handleFloatField,
		fieldHumidity:    pkt.handleIntField,
		fieldSubtype:     pkt.handleByteField,
		fieldEvents:      pkt.handleByteField,
		fieldAccel:       pkt.handleAccelField,
		fieldAccels:      pkt.handleAccelsField,
		fieldRange:       pkt.handleIntField,
		fieldCO2:         pkt.handleUintField,
		fieldCounter:     pkt.handleUintField,
		fieldUserData:    pkt.handleIntField,
		fieldReserved:    pkt.handleReservedField,
		fieldReserved2:   pkt.handleReserved2Field,
		fieldGP:          pkt.handleGpField,
		fieldBattAct:     pkt.handleBattActField,
		fieldRsEvents:    pkt.handleRsEventsField,
	}

	if model, ok := def.model.(string); ok {
		pkt.msdata[fieldModel] = model
	} else if meth, ok := def.model.(func() string); ok {
		pkt.msdata[fieldModel] = meth()
	}
	index := 4
	for _, fieldName := range def.fields {
		if fieldMethod, ok := fieldDefs[fieldName]; ok {
			index = fieldMethod(fieldName, index)
		}
	}
	if value, ok := pkt.msdata[fieldEvents]; ok && len(def.events) > 0 {
		for _, evt := range def.events {
			pkt.msdata[evt] = value.(uint8)&evtMasks[evt] != 0
		}
	}
	return true
}

// Parse the payload by checkout 'subtype' first
// Will find definition by subtype in the input 'payloadDefs'
func (pkt Payload) parsePayloadBySubtype(payloadDefs map[byte]payloadDef) bool {
	msd := pkt.ManufacturerData()
	pkt.msdata[fieldVendor] = knownVendorCode[IngicsVendorCode]
	subtype := uint8(msd[13])
	if def, ok := payloadDefs[subtype]; ok {
		return pkt.parsePayload(def)
	}
	return false
}
