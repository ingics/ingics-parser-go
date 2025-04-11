package ibs

import (
	"encoding/binary"
)

const IngicsVendorCode = 0x082C

type AccelReading struct {
	X int16
	Y int16
	Z int16
}

const (
	fieldModel       = "model"
	fieldVendor      = "vendor"
	fieldBattery     = "battery"
	fieldTemperature = "temperature"
	fieldHumidity    = "humidity"   // 1% resolution (integer) most cases
	fieldHumidity1D  = "humidity1D" // 0.1% resolution for iWS01/iBS08T
	fieldTempExt     = "temperatureExt"
	fieldTempEnv     = "temperatureEnv"
	fieldRange       = "range"
	fieldGP          = "gp"
	fieldCounter     = "counter"
	fieldCO2         = "co2"
	fieldAccel       = "accel"
	fieldAccels      = "accels"
	fieldLux         = "lux"
	fieldUserData    = "userdata"
	fieldEvents      = "events"
	fieldSubtype     = "subtype"
	fieldReserved    = "reserved"
	fieldReserved2   = "reserved2"
	fieldBattAct     = "battact"
	fieldRsEvents    = "rsEvents"
	fieldVoltage     = "voltage"
	fieldCurrent     = "current"
	fieldValue       = "value"
	fieldPm2p5       = "pm2p5"
	fieldPm10p0      = "pm10p0"
	fieldVoc         = "voc"
	fieldNox         = "nox"
	fieldAux1        = "aux1"
	fieldAux2        = "aux2"
	fieldAux3        = "aux3"
)

const (
	evtButton = "button"
	evtMoving = "moving"
	evtHall   = "hall"
	evtFall   = "fall"
	evtPIR    = "pir"
	evtIR     = "ir"
	evtDetect = "detect"
	evtDin    = "din"
	evtDin2   = "din2"
	evtFlip   = "flip"
	bitButton = 0
	bitMoving = 1
	bitHall   = 2
	bitFall   = 3
	bitPIR    = 4
	bitIR     = 5
	bitDetect = 5 // for iBS08
	bitDin    = 6
	bitDin2   = 3 // for iBS03QY
	bitFlip   = 5 // for iBs05G_Flip
)

var evtMasks = map[string]uint8{
	evtButton: uint8(1 << bitButton),
	evtMoving: uint8(1 << bitMoving),
	evtHall:   uint8(1 << bitHall),
	evtFall:   uint8(1 << bitFall),
	evtPIR:    uint8(1 << bitPIR),
	evtIR:     uint8(1 << bitIR),
	evtDetect: uint8(1 << bitDetect),
	evtDin:    uint8(1 << bitDin),
	evtDin2:   uint8(1 << bitDin2),
	evtFlip:   uint8(1 << bitFlip),
}

func (pkt Payload) ibs() bool {
	if mfg, ok := pkt.VendorCode(); ok {
		msd := pkt.Packet.ManufacturerData()
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
			return pkt.parsePayloadBySubtype(13, rsPayloadDefs)
		} else if mfg == 0x0D && code == 0xBC83 {
			// iBS02/iBS03/iBS04 common payload
			return pkt.parsePayloadBySubtype(13, ibsCommonPayloadDefs)
		} else if mfg == 0x0D && code == 0x0BC85 {
			// iBS03GP
			var gpPayloadDef = payloadDef{
				"iBS03GP",
				[]string{fieldBattAct, fieldAccels, fieldGP},
				[]string{},
			}
			return pkt.parsePayload(gpPayloadDef)
		} else if mfg == 0x082C && code == 0x0BC86 {
			// iBS05RG
			var gpPayloadDef = payloadDef{
				"iBS05RG",
				[]string{fieldBattAct, fieldAccels},
				[]string{},
			}
			return pkt.parsePayload(gpPayloadDef)
		} else if mfg == 0x082C && code == 0xBC83 {
			// iBS05/iBS06
			return pkt.parsePayloadBySubtype(13, ibsCommonPayloadDefs)
		} else if mfg == 0x082C && code == 0xBC87 {
			// iBS07/iWS01
			return pkt.parsePayloadBySubtype(19, ibsBC87PayloadDefs)
		} else if mfg == 0x082C && code == 0xBC88 {
			// iBS08/iBS09
			return pkt.parsePayloadBySubtype(21, ibsBC88PayloadDefs)
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
				return pkt.parsePayloadBySubtype(13, ibs01PayloadDefs)
			}
		}
	}
	return false
}

func (pkt Payload) handleIntField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xAAAA {
		pkt.msdata[name] = int16(unsignedValue)
		// special handling for humidity, convert to float
		if name == fieldHumidity {
			pkt.msdata[name] = float32(pkt.msdata[name].(int16))
		}
	}
	return index + 2
}

func (pkt Payload) handleUserDataField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	pkt.msdata[name] = int16(unsignedValue)
	return index + 2
}

func (pkt Payload) handleUintField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xFFFF {
		pkt.msdata[name] = unsignedValue
	}
	return index + 2
}

func (pkt Payload) handleFloatField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xAAAA {
		pkt.msdata[name] = float32(int16(unsignedValue)) / 100.0
	}
	return index + 2
}

func (pkt Payload) handleHumidity(name string, index int) int {
	pkt.handleUintField(name, index)
	if v, ok := pkt.msdata[name].(uint16); ok {
		pkt.msdata[name] = float32(int16(v))
	} else {
		delete(pkt.msdata, name)
	}
	return index + 2
}

func (pkt Payload) handleHumidity1D(name string, index int) int {
	pkt.handleUintField(fieldHumidity, index)
	if v, ok := pkt.msdata[fieldHumidity].(uint16); ok {
		pkt.msdata[fieldHumidity] = float32(int16(v)) / 10.0
	} else {
		delete(pkt.msdata, fieldHumidity)
	}
	return index + 2
}

func (pkt Payload) handleUint1DField(name string, index int) int {
	msd := pkt.ManufacturerData()
	unsignedValue := binary.LittleEndian.Uint16(msd[index : index+2])
	if unsignedValue != 0xFFFF {
		pkt.msdata[name] = float32(int16(unsignedValue)) / 10.0
	}
	return index + 2
}

func (pkt Payload) handleAccelField(name string, index int) int {
	msd := pkt.ManufacturerData()
	pkt.msdata[name] = AccelReading{
		int16(binary.LittleEndian.Uint16(msd[index : index+2])),
		int16(binary.LittleEndian.Uint16(msd[index+2 : index+4])),
		int16(binary.LittleEndian.Uint16(msd[index+4 : index+6])),
	}
	return index + 6
}

func (pkt Payload) handleAccelsField(name string, index int) int {
	msd := pkt.ManufacturerData()
	pkt.msdata[name] = []AccelReading{
		{
			int16(binary.LittleEndian.Uint16(msd[index : index+2])),
			int16(binary.LittleEndian.Uint16(msd[index+2 : index+4])),
			int16(binary.LittleEndian.Uint16(msd[index+4 : index+6])),
		},
		{
			int16(binary.LittleEndian.Uint16(msd[index+6 : index+8])),
			int16(binary.LittleEndian.Uint16(msd[index+8 : index+10])),
			int16(binary.LittleEndian.Uint16(msd[index+10 : index+12])),
		},
		{
			int16(binary.LittleEndian.Uint16(msd[index+12 : index+14])),
			int16(binary.LittleEndian.Uint16(msd[index+14 : index+16])),
			int16(binary.LittleEndian.Uint16(msd[index+16 : index+18])),
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
	pkt.msdata[name] = float32(unsignedValue) / 50.0
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
// returns model name determine by mfg code
func (pkt Payload) rgNaming() string {
	if mfg, ok := pkt.VendorCode(); ok {
		if mfg == 0x59 {
			return "iBS01RG"
		} else if mfg == 0x0D {
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
		"iBS02PIR2-RS",
		[]string{fieldBattery, fieldRsEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
	0x02: {
		"iBS02IR2-RS",
		[]string{fieldBattery, fieldRsEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
	0x04: {
		"iBS02M2-RS",
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
	0x1A: {
		"iBS03RS",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldRange, fieldUserData},
		[]string{},
	},
	0x1B: {
		"iBS03F",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtDin},
	},
	0x1C: {
		"iBS03Q",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtDin},
	},
	0x1D: {
		"iBS03QY",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtDin, evtDin2},
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
	0x23: {
		"iBS03AD-NTC",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldTempExt, fieldUserData},
		[]string{},
	},
	0x24: {
		"iBS03AD-V",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldVoltage, fieldUserData},
		[]string{},
	},
	0x25: {
		"iBS03AD-D",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
		[]string{evtDin},
	},
	0x26: {
		"iBS03AD-A",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldCurrent, fieldUserData},
		[]string{},
	},
	0x30: {
		"iBS05",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x31: {
		"iBS05H",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldCounter, fieldUserData},
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
	0x35: {
		"iBS05i",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x36: {
		"iBS06i",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton},
	},
	0x3A: {
		"iBS05G-Flip",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{evtButton, evtFlip},
	},
	0x40: {
		"iBS06",
		[]string{fieldBattery, fieldReserved, fieldReserved2, fieldReserved2, fieldUserData},
		[]string{},
	},
}

// product ID BC87
var ibsBC87PayloadDefs = map[byte]payloadDef{
	0x50: {
		"iBS07",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldHumidity, fieldLux, fieldAccel},
		[]string{evtButton},
	},
}

// product ID BC88
var ibsBC88PayloadDefs = map[byte]payloadDef{
	0x42: {
		"iBS09R",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldRange, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtButton, evtDetect},
	},
	0x43: {
		"iBS09PS",
		[]string{fieldBattery, fieldEvents, fieldValue, fieldCounter, fieldReserved2, fieldAux1, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtDetect},
	},
	0x44: {
		"iBS09PIR",
		[]string{fieldBattery, fieldEvents, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtPIR},
	},
	0x45: {
		"iBS08T",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldHumidity1D, fieldLux, fieldReserved2, fieldReserved2, fieldReserved2, fieldReserved2},
		[]string{evtButton},
	},
	0x46: {
		"iBS08IAQ",
		[]string{fieldBattery, fieldEvents, fieldTemperature, fieldHumidity1D, fieldCO2, fieldPm2p5, fieldPm10p0, fieldVoc, fieldNox},
		[]string{evtButton},
	},
}

// Parse the payload follow the input definition
func (pkt Payload) parsePayload(def payloadDef) bool {
	// define field handlers
	pkt.msdata[fieldVendor] = knownVendorCode[IngicsVendorCode]
	var fieldDefs = map[string]func(string, int) int{
		fieldBattery:     pkt.handleFloatField,
		fieldTemperature: pkt.handleFloatField,
		fieldTempExt:     pkt.handleFloatField,
		fieldTempEnv:     pkt.handleFloatField,
		fieldHumidity:    pkt.handleHumidity,
		fieldHumidity1D:  pkt.handleHumidity1D,
		fieldSubtype:     pkt.handleByteField,
		fieldEvents:      pkt.handleByteField,
		fieldAccel:       pkt.handleAccelField,
		fieldAccels:      pkt.handleAccelsField,
		fieldRange:       pkt.handleIntField,
		fieldCO2:         pkt.handleUintField,
		fieldCounter:     pkt.handleUintField,
		fieldUserData:    pkt.handleUserDataField,
		fieldReserved:    pkt.handleReservedField,
		fieldReserved2:   pkt.handleReserved2Field,
		fieldGP:          pkt.handleGpField,
		fieldBattAct:     pkt.handleBattActField,
		fieldRsEvents:    pkt.handleRsEventsField,
		fieldLux:         pkt.handleUintField,
		fieldVoltage:     pkt.handleIntField,
		fieldCurrent:     pkt.handleUintField,
		fieldValue:       pkt.handleIntField,
		fieldPm2p5:       pkt.handleUint1DField,
		fieldPm10p0:      pkt.handleUint1DField,
		fieldVoc:         pkt.handleUint1DField,
		fieldNox:         pkt.handleUint1DField,
		fieldAux1:        pkt.handleIntField,
		fieldAux2:        pkt.handleIntField,
		fieldAux3:        pkt.handleIntField,
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
func (pkt Payload) parsePayloadBySubtype(subTypeIdx int, payloadDefs map[byte]payloadDef) bool {
	msd := pkt.ManufacturerData()
	pkt.msdata[fieldVendor] = knownVendorCode[IngicsVendorCode]
	subtype := uint8(msd[subTypeIdx])
	if def, ok := payloadDefs[subtype]; ok {
		return pkt.parsePayload(def)
	}
	return false
}
