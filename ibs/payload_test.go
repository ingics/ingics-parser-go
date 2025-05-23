package ibs

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/adv"
	"github.com/google/uuid"
)

func validateFieldFunc(t *testing.T, got *Payload, field string, want interface{}) {
	t.Run(field, func(t *testing.T) {
		if meth := reflect.ValueOf(got).MethodByName(field); meth.IsValid() {
			out := meth.Call(nil)
			if ok := out[1].Bool(); !ok {
				if want != nil {
					t.Errorf("got.%v call failed", field)
				}
				return
			}
			value := out[0].Interface()
			if _, ok := value.(string); ok {
				// trim invisible (non-printable) characters for string field
				value = strings.TrimFunc(value.(string), func(r rune) bool {
					// trim invisible (non-printable) characters
					return !unicode.IsGraphic(r)
				})
			}
			if !reflect.DeepEqual(value, want) {
				t.Errorf(
					"got.%v = %v (%v), want = %v (%v)",
					field, value, reflect.TypeOf(value), want, reflect.TypeOf(want))
			}
		} else {
			t.Errorf("got.%v not found", field)
		}
	})
}

type TestCaseField struct {
	name   string
	expect interface{}
}

type TestCase struct {
	payload string
	fields  []TestCaseField
}

func runTestCases(t *testing.T, cases []TestCase) {
	for _, v := range cases {
		payload, _ := hex.DecodeString(v.payload)
		got := Parse(payload)
		for _, f := range v.fields {
			validateFieldFunc(t, got, f.name, f.expect)
		}
	}
}

func TestParse_Windows10(t *testing.T) {
	payload, _ := hex.DecodeString("1EFF06000109200236444DA103B7448CE1A6E2220F1E9AB734C9348A35B53B")
	got := Parse(payload)
	validateFieldFunc(t, got, "Vendor", "Microsoft")
	validateFieldFunc(t, got, "ProductModel", "Windows 10 Desktop")
}

func TestParse_IBeacon(t *testing.T) {
	payload, _ := hex.DecodeString("0201061AFF4C000215E2C56DB5DFFB48D2B060D0F5A71096E000000000C5")
	got := Parse(payload)
	t.Run("Packet", func(t *testing.T) {
		uuid, _ := ble.Parse("E2C56DB5-DFFB-48D2-B060-D0F5A71096E0")
		want, _ := adv.NewPacket(adv.Flags(6), adv.IBeacon(uuid, 0, 0, -59))
		if !reflect.DeepEqual(&got.Packet, want) {
			t.Errorf("adv.Packet = %v, want %v", &got.Packet, want)
		}
	})
	validateFieldFunc(t, got, "Vendor", "Apple, Inc.")
	validateFieldFunc(t, got, "ProductModel", "iBeacon")
}

func TestParse_UnknownVendor(t *testing.T) {
	payload, _ := hex.DecodeString("0201061AFFF0080215E2C56DB5DFFB48D2B060D0F5A71096E000000000C5")
	got := Parse(payload)
	validateFieldFunc(t, got, "Vendor", "0x08F0")
	validateFieldFunc(t, got, "ProductModel", nil)
}

func TestParse_IBeaconResp(t *testing.T) {
	payload, _ := hex.DecodeString("020A000816F0FF640000000012094D696E69426561636F6E5F303731343700")
	got := Parse(payload)
	validateFieldFunc(t, got, "LocalName", "MiniBeacon_07147")
	validateFieldFunc(t, got, "Vendor", nil)
	validateFieldFunc(t, got, "ProductModel", nil)
	t.Run("ServiceData", func(t *testing.T) {
		wantdata, _ := hex.DecodeString("6400000000")
		found := false
		for _, data := range got.ServiceData() {
			if data.UUID.Equal(ble.UUID16(0xFFF0)) && reflect.DeepEqual(data.Data, wantdata) {
				found = true
				break
			}
		}
		if !found {
			servicedata := ble.ServiceData{UUID: ble.UUID16(0xFFF0), Data: wantdata}
			t.Errorf("got.ServiceData not contain %v\ngot.ServiceData() = %v", servicedata, got.ServiceData())
		}
	})
}

func TestParse_IBS01_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC360101FFFFFFFFFFFFFFFFFFFF")
	got := Parse(payload)
	validateFieldFunc(t, got, "Vendor", "INGICS TECHNOLOGY CO., LTD.")
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.10))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

func TestParse_IBS01T_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BCFF00007A0D4300FFFFFFFFFFFF")
	got := Parse(payload)
	validateFieldFunc(t, got, "Vendor", "INGICS TECHNOLOGY CO., LTD.")
	validateFieldFunc(t, got, "ProductModel", "iBS01T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.55))
	validateFieldFunc(t, got, "Temperature", float32(34.50))
	validateFieldFunc(t, got, "Humidity", float32(67))
}

func TestParse_IBS01H_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2B0104FFFFFFFFFFFFFFFFFFFF")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.99))
	validateFieldFunc(t, got, "HallDetected", true)
}

func TestParse_IBS01G_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2B010AFFFFFFFFFFFFFFFFFFFF")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.99))
	validateFieldFunc(t, got, "Falling", true)
	validateFieldFunc(t, got, "Moving", true)
}

func TestParse_IBS01T_New(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2E0100BFFA3900000005000000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.02))
	validateFieldFunc(t, got, "Temperature", float32(-13.45))
	validateFieldFunc(t, got, "Humidity", float32(57))
}

func TestParse_IBS02IR(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC200120AAAAFFFF000002070000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02IR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.88))
	validateFieldFunc(t, got, "Counter", nil)
	validateFieldFunc(t, got, "IRDetected", true)
}

func TestParse_IBS02IR_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4D0120AAAA05000000020A0600")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02IR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.33))
	validateFieldFunc(t, got, "Counter", 5)
	validateFieldFunc(t, got, "IRDetected", true)
}

func TestParse_IBS02PIR_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4A0110AAAAFFFF000001140000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02PIR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
	validateFieldFunc(t, got, "ButtonPressed", nil)
	validateFieldFunc(t, got, "PIRDetected", true)
}

func TestParse_iBS02M2(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC3E0140AAAAFFFF000004070000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02M2")
	validateFieldFunc(t, got, "Counter", nil)
	validateFieldFunc(t, got, "IRDetected", nil)
	validateFieldFunc(t, got, "DinTriggered", true)
	validateFieldFunc(t, got, "ButtonPressed", nil)
}

func TestParse_iBS02M2_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC240100AAAA37060000040B0600")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02M2")
	validateFieldFunc(t, got, "Counter", 1591)
	validateFieldFunc(t, got, "IRDetected", nil)
	validateFieldFunc(t, got, "DinTriggered", false)
	validateFieldFunc(t, got, "ButtonPressed", nil)
}

func TestParse_IBS03T(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC2801020A09FFFF000015030000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", float32(23.14))
	validateFieldFunc(t, got, "Humidity", nil)
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParse_IBS03T_RH(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BCAD0000A20B4700FFFF14000000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03T")
	validateFieldFunc(t, got, "Temperature", float32(29.78))
	validateFieldFunc(t, got, "Humidity", float32(71))
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParse_IBS03RG(t *testing.T) {
	payload, _ := hex.DecodeString("02010619FF0D0081BC3E110A00F4FF00FF1600F6FF00FF1400F6FF08FF")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03RG")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.18))
	validateFieldFunc(t, got, "Moving", true)
	validateFieldFunc(t, got, "Accels", []AccelReading{{10, -12, -256}, {22, -10, -256}, {20, -10, -248}})
}

func TestParse_IBS05RG(t *testing.T) {
	payload, _ := hex.DecodeString("0201061BFF2C0886BC3E110A00F4FF00FF1600F6FF00FF1400F6FF08FF1704")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05RG")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.18))
	validateFieldFunc(t, got, "Moving", true)
	validateFieldFunc(t, got, "Accels", []AccelReading{{10, -12, -256}, {22, -10, -256}, {20, -10, -248}})
}

func TestParse_IBS03TP(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100D809060A640017040000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03TP")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", float32(25.20))
	validateFieldFunc(t, got, "TemperatureExt", float32(25.66))
}

func TestParse_IBS03R(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100AAAA7200000013090000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03R")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Range", 114)
	validateFieldFunc(t, got, "Temperature", nil)
}

func TestParse_IBS03RS(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC430100AAAA150000001A040600")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03RS")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.23))
	validateFieldFunc(t, got, "Range", 21)
}

func TestParse_IBS03P(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC2C0100BF0AD00A0000120A0600")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03P")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3))
	validateFieldFunc(t, got, "Range", nil)
	validateFieldFunc(t, got, "Temperature", float32(27.51))
	validateFieldFunc(t, got, "TemperatureExt", float32(27.68))
	validateFieldFunc(t, got, "Humidity", nil)
}

func TestParse_IBS03GP(t *testing.T) {
	payload, _ := hex.DecodeString("0201061BFF0D0085BC3111160082FF9EFE4E001200D2FE10003A005CFFD9C5")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03GP")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.05))
	validateFieldFunc(t, got, "Accels", []AccelReading{{22, -126, -354}, {78, 18, -302}, {16, 58, -164}})
	validateFieldFunc(t, got, "Moving", true)
	validateFieldFunc(t, got, "GP", float32(1012.98))
}

func TestParse_IBS03F(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC290140AAAA020000001B090000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03F")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "Counter", 2)
	validateFieldFunc(t, got, "DinTriggered", true)
}

func TestParse_IBS03Q(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC290140AAAA000000001C090000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03Q")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "Counter", 0)
	validateFieldFunc(t, got, "DinTriggered", true)
}

func TestParse_IBS03QY(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC330140AAAA0A0000001D090000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03QY")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.07))
	validateFieldFunc(t, got, "Counter", 10)
	validateFieldFunc(t, got, "DinTriggered", true)
	payload2, _ := hex.DecodeString("02010612FF0D0083BC330108AAAA0A0000001D090000")
	got2 := Parse(payload2)
	validateFieldFunc(t, got2, "ProductModel", "iBS03QY")
	validateFieldFunc(t, got2, "BatteryVoltage", float32(3.07))
	validateFieldFunc(t, got2, "Counter", 10)
	validateFieldFunc(t, got2, "DinTriggered", false)
	validateFieldFunc(t, got2, "Din2Triggered", true)
}

func TestParse_IBS04(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC3A0101AAAAFFFF000019070000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS04")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.14))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

// test SCAN RESPONSE of iBS04i
func TestParse_IBS04i(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC1F0100AAAAFFFF000018030000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS04i")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.87))
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParse_IBS05(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC290101AAAAFFFF000030000000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

func TestParse_IBS05H(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC2D0100AAAA04000000310A1000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05H")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.01))
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "HallDetected", false)
	payload, _ = hex.DecodeString("02010612FF2C0883BC2D0101AAAA04000000310A1000")
	got = Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05H")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.01))
	validateFieldFunc(t, got, "ButtonPressed", true)
	validateFieldFunc(t, got, "HallDetected", false)
	payload, _ = hex.DecodeString("02010612FF2C0883BC2D0104AAAA01800000310A1000")
	got = Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05H")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.01))
	validateFieldFunc(t, got, "Counter", 32769)
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "HallDetected", true)
}

func TestParse_IBS05T(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC4A0100A10AFFFF000032000000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "Temperature", float32(27.21))
	validateFieldFunc(t, got, "Humidity", nil)
}

func TestParse_IBS05G(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC290102AAAAFFFF000033000000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05G")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "Moving", true)
}

func TestParse_IBS05G_Flip(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"02010612FF2C0883BC3C012002FF000000003A0A1000",
			[]TestCaseField{
				{"ProductModel", "iBS05G-Flip"},
				{"BatteryVoltage", float32(3.16)},
				{"ButtonPressed", false},
				{"Flip", true},
			},
		},
		{
			"02010612FF2C0883BC3A0101F200000000003A0A1000",
			[]TestCaseField{
				{"ProductModel", "iBS05G-Flip"},
				{"BatteryVoltage", float32(3.14)},
				{"ButtonPressed", true},
				{"Flip", false},
			},
		},
	})
}

func TestParse_IBS05CO2(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC270100AAAA6804000034010000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05CO2")
	validateFieldFunc(t, got, "Temperature", nil)
	validateFieldFunc(t, got, "CO2", 1128)
}

func TestParse_IBS06(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC4A0100AAAAFFFF000040110000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS06")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
}

func TestParse_IBS02HM(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0082BC280100AAAAFFFF000004050000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02M2-RS")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
}

func TestParse_IRS02RG(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4D010000002400FCFE22074B58")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iRS02RG")
	validateFieldFunc(t, got, "Accel", AccelReading{0, 36, -260})
}

func TestParse_CfgService(t *testing.T) {
	uuid := []byte{0x2B, 0x32, 0x64, 0xB4, 0x1C, 0x6D, 0x1A, 0x84, 0xBD, 0x46, 0x98, 0xB2, 0x00, 0x00, 0x4E, 0x1B}
	payload, _ := hex.DecodeString("11072B3264B41C6D1A84BD4698B200004E1B0B0969425330352D44384242")
	got := Parse(payload)
	// fmt.Printf("%s\n", got.Packet.UUIDs())
	if !ble.Contains(got.Packet.UUIDs(), uuid) {
		t.Errorf("INGICS tag configuration service not found")
	}
}

func TestParse_IBS07(t *testing.T) {
	payload, _ := hex.DecodeString("02010618FF2C0887BC330100110B31005A002AFF02007B0050070000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS07")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.07))
	validateFieldFunc(t, got, "Temperature", float32(28.33))
	validateFieldFunc(t, got, "Humidity", float32(49))
	validateFieldFunc(t, got, "Lux", uint(90))
	validateFieldFunc(t, got, "Accel", AccelReading{-214, 2, 123})
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParse_IBS07_NoSensor(t *testing.T) {
	payload, _ := hex.DecodeString("02010618FF2C0887BC330101AAAAFFFF00002AFF02007B0050070000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS07")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.07))
	validateFieldFunc(t, got, "Temperature", nil)
	validateFieldFunc(t, got, "Humidity", nil)
	validateFieldFunc(t, got, "Lux", uint(0))
	validateFieldFunc(t, got, "Accel", AccelReading{-214, 2, 123})
	validateFieldFunc(t, got, "ButtonPressed", true)
}

func TestParse_IBS03AD_NTC(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100AAAA060A640023040000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03AD-NTC")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", nil)
	validateFieldFunc(t, got, "TemperatureExt", float32(25.66))
	validateFieldFunc(t, got, "UserData", int(100))
}

func TestParse_IBS03AD_V(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100AAAA060A640024040000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03AD-V")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", nil)
	validateFieldFunc(t, got, "Voltage", 2566)
	validateFieldFunc(t, got, "UserData", int(100))
}

func TestParse_IBS03AD_D(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280140AAAA060A640025040000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03AD-D")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", nil)
	validateFieldFunc(t, got, "Voltage", nil)
	validateFieldFunc(t, got, "DinTriggered", true)
	validateFieldFunc(t, got, "UserData", int(100))
	validateFieldFunc(t, got, "Counter", int(2566))
}

func TestParse_IBS03AD_A(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280140D809060A640026040000")
	got := Parse(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03AD-A")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Current", uint(2566))
	validateFieldFunc(t, got, "UserData", int(100))
}

func TestParse_IBS09R(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC470100AAAA74000000000000000000000042100000",
			[]TestCaseField{
				{"ProductModel", "iBS09R"},
				{"BatteryVoltage", float32(3.27)},
				{"Range", int(116)},
				{"Detected", false},
			},
		},
		{
			"0201061AFF2C0888BC470120AAAA74000000000000000000000042100000",
			[]TestCaseField{
				{"ProductModel", "iBS09R"},
				{"BatteryVoltage", float32(3.27)},
				{"Range", int(116)},
				{"Detected", true},
			},
		},
	})
}

func TestParse_IBS09PS(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC470120AAAA01000000000000000000000043100000",
			[]TestCaseField{
				{"ProductModel", "iBS09PS"},
				{"BatteryVoltage", float32(3.27)},
				{"Counter", int(1)},
				{"Detected", true},
			},
		},
		{
			"0201061AFF2C0888BC470100AAAA00000000000000000000000043100000",
			[]TestCaseField{
				{"ProductModel", "iBS09PS"},
				{"BatteryVoltage", float32(3.27)},
				{"Counter", int(0)},
				{"Detected", false},
			},
		},
	})
}

func TestParse_IBS09PIR(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC470110AAAAFFFF0000000000000000000044100000",
			[]TestCaseField{
				{"ProductModel", "iBS09PIR"},
				{"BatteryVoltage", float32(3.27)},
				{"ButtonPressed", nil},
				{"PIRDetected", true},
			},
		},
		{
			"0201061AFF2C0888BCFA0000AAAAFFFF0000000000000000000044100000",
			[]TestCaseField{
				{"ProductModel", "iBS09PIR"},
				{"BatteryVoltage", float32(2.50)},
				{"ButtonPressed", nil},
				{"PIRDetected", false},
			},
		},
	})
}

func TestParse_IBS08T(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC4701010B0BA3010102000000000000000045100000",
			[]TestCaseField{
				{"ProductModel", "iBS08T"},
				{"BatteryVoltage", float32(3.27)},
				{"Temperature", float32(28.27)},
				{"Humidity", float32(41.9)},
				{"Lux", uint(513)},
				{"ButtonPressed", true},
			},
		},
		{
			"0201061AFF2C0888BC2001006E281300B907000000000000000045100000",
			[]TestCaseField{
				{"ProductModel", "iBS08T"},
				{"BatteryVoltage", float32(2.88)},
				{"Temperature", float32(103.5)},
				{"Humidity", float32(1.9)},
				{"Lux", uint(1977)},
				{"ButtonPressed", false},
			},
		},
		{
			"0201061AFF2C0888BC330100870DF5019982000000000000000045020900",
			[]TestCaseField{
				{"ProductModel", "iBS08T"},
				{"BatteryVoltage", float32(3.07)},
				{"Temperature", float32(34.63)},
				{"Humidity", float32(50.1)},
				{"Lux", uint(33433)},
				{"ButtonPressed", false},
			},
		},
	})
}

func TestParse_IBS08IAQ(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC4901000F091F025A0232004C00DE030A0046040000",
			[]TestCaseField{
				{"ProductModel", "iBS08IAQ"},
				{"BatteryVoltage", float32(3.29)},
				{"Temperature", float32(23.19)},
				{"Humidity", float32(54.3)},
				{"CO2", 602},
				{"PM2p5", float32(5.0)},
				{"PM10p0", float32(7.6)},
				{"VOC", float32(99.0)},
				{"NOx", float32(1.0)},
				{"ButtonPressed", false},
			},
		},
		{
			"0201061AFF2C0888BC4701001A091C02FFFFFFFFFFFF0000FFFF46041000",
			[]TestCaseField{
				{"ProductModel", "iBS08IAQ"},
				{"BatteryVoltage", float32(3.27)},
				{"Temperature", float32(23.30)},
				{"Humidity", float32(54.0)},
				{"CO2", nil},
				{"PM2p5", nil},
				{"PM10p0", nil},
				{"VOC", float32(0)},
				{"NOx", nil},
				{"ButtonPressed", false},
			},
		},
	})
}

func TestParse_IBS09IR(t *testing.T) {
	runTestCases(t, []TestCase{
		{
			"0201061AFF2C0888BC390120AAAAD1000000060000000000000047080000",
			[]TestCaseField{
				{"ProductModel", "iBS09IR"},
				{"BatteryVoltage", float32(3.13)},
				{"ButtonPressed", false},
				{"IRDetected", true},
				{"Counter", 209},
			},
		},
		{
			"0201061AFF2C0888BC390101AAAAD0000000040000000000000047080000",
			[]TestCaseField{
				{"ProductModel", "iBS09IR"},
				{"BatteryVoltage", float32(3.13)},
				{"ButtonPressed", true},
				{"IRDetected", false},
				{"Counter", 208},
			},
		},
	})
}

func TestParser_iBeacon(t *testing.T) {
	ProximityUUIDs, _ := uuid.MustParse("B9A5D27D56CC4E3AAB511F2153BCB967").MarshalBinary()
	runTestCases(t, []TestCase{
		{
			"0201061AFF4C000215B9A5D27D56CC4E3AAB511F2153BCB9670000E9B2D6",
			[]TestCaseField{
				{"Vendor", "Apple, Inc."},
				{"ProductModel", "iBeacon"},
				{"Major", uint(0)},
				{"Minor", uint(0xE9B2)},
				{"RefTx", int(-42)},
				{"UUID", ProximityUUIDs},
			},
		},
	})
}
