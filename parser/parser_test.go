package parser

import (
	"encoding/hex"
	"reflect"
	"strings"
	"testing"
	"unicode"

	"github.com/go-ble/ble"
	"github.com/go-ble/ble/linux/adv"
)

func validateFieldFunc(t *testing.T, got *Packet, field string, want interface{}) {
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

func TestParseBlePayload_Windows10(t *testing.T) {
	payload, _ := hex.DecodeString("1EFF06000109200236444DA103B7448CE1A6E2220F1E9AB734C9348A35B53B")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "Vendor", "Microsoft")
	validateFieldFunc(t, got, "ProductModel", "Windows 10 Desktop")
}

func TestParseBlePayload_IBeacon(t *testing.T) {
	payload, _ := hex.DecodeString("0201061AFF4C000215E2C56DB5DFFB48D2B060D0F5A71096E000000000C5")
	got := ParseBlePayload(payload)
	t.Run("Packet", func(t *testing.T) {
		uuid, _ := ble.Parse("E2C56DB5-DFFB-48D2-B060-D0F5A71096E0")
		want, _ := adv.NewPacket(adv.Flags(6), adv.IBeacon(uuid, 0, 0, -59))
		if !reflect.DeepEqual(&got.adv, want) {
			t.Errorf("adv.Packet = %v, want %v", &got.adv, want)
		}
	})
	validateFieldFunc(t, got, "Vendor", "Apple, Inc.")
	validateFieldFunc(t, got, "ProductModel", "iBeacon")
}

func TestParseBlePayload_UnknownVendor(t *testing.T) {
	payload, _ := hex.DecodeString("0201061AFFF0080215E2C56DB5DFFB48D2B060D0F5A71096E000000000C5")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "Vendor", "0x08F0")
	validateFieldFunc(t, got, "ProductModel", nil)
}

func TestParseBlePayload_IBeaconResp(t *testing.T) {
	payload, _ := hex.DecodeString("020A000816F0FF640000000012094D696E69426561636F6E5F303731343700")
	got := ParseBlePayload(payload)
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

func TestParseBlePayload_IBS01_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC360101FFFFFFFFFFFFFFFFFFFF")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "Vendor", "INGICS TECHNOLOGY CO., LTD.")
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.10))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

func TestParseBlePayload_IBS01T_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BCFF00007A0D4300FFFFFFFFFFFF")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "Vendor", "INGICS TECHNOLOGY CO., LTD.")
	validateFieldFunc(t, got, "ProductModel", "iBS01T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.55))
	validateFieldFunc(t, got, "Temperature", float32(34.50))
	validateFieldFunc(t, got, "Humidity", int16(67))
}

func TestParseBlePayload_IBS01H_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2B0104FFFFFFFFFFFFFFFFFFFF")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.99))
	validateFieldFunc(t, got, "HallDetected", true)
}

func TestParseBlePayload_IBS01G_Old(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2B010AFFFFFFFFFFFFFFFFFFFF")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.99))
	validateFieldFunc(t, got, "Falling", true)
	validateFieldFunc(t, got, "Moving", true)
}

func TestParseBlePayload_IBS01T_New(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF590080BC2E0100BFFA3900000005000000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS01T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.02))
	validateFieldFunc(t, got, "Temperature", float32(-13.45))
	validateFieldFunc(t, got, "Humidity", int16(57))
}

func TestParseBlePayload_IBS02IR(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC200120AAAAFFFF000002070000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02IR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.88))
	validateFieldFunc(t, got, "Counter", nil)
	validateFieldFunc(t, got, "IRDetected", true)
}

func TestParseBlePayload_IBS02IR_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4D0120AAAA05000000020A0600")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02IR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.33))
	validateFieldFunc(t, got, "Counter", uint16(5))
	validateFieldFunc(t, got, "IRDetected", true)
}

func TestParseBlePayload_IBS02PIR_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4A0110AAAAFFFF000001140000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02PIR2")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
	validateFieldFunc(t, got, "ButtonPressed", nil)
	validateFieldFunc(t, got, "PIRDetected", true)
}

func TestParseBlePayload_iBS02M2(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC3E0140AAAAFFFF000004070000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02M2")
	validateFieldFunc(t, got, "Counter", nil)
	validateFieldFunc(t, got, "IRDetected", nil)
	validateFieldFunc(t, got, "DinTriggered", true)
	validateFieldFunc(t, got, "ButtonPressed", nil)
}

func TestParseBlePayload_iBS02M2_counter(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC240100AAAA37060000040B0600")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02M2")
	validateFieldFunc(t, got, "Counter", uint16(1591))
	validateFieldFunc(t, got, "IRDetected", nil)
	validateFieldFunc(t, got, "DinTriggered", false)
	validateFieldFunc(t, got, "ButtonPressed", nil)
}

func TestParseBlePayload_IBS03T(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC2801020A09FFFF000015030000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", float32(23.14))
	validateFieldFunc(t, got, "Humidity", nil)
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParseBlePayload_IBS03T_RH(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BCAD0000A20B4700FFFF14000000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03T")
	validateFieldFunc(t, got, "Temperature", float32(29.78))
	validateFieldFunc(t, got, "Humidity", int16(71))
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParseBlePayload_IBS03RG(t *testing.T) {
	payload, _ := hex.DecodeString("02010619FF0D0081BC3E110A00F4FF00FF1600F6FF00FF1400F6FF08FF")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03RG")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.18))
	validateFieldFunc(t, got, "Moving", true)
	validateFieldFunc(t, got, "Accels", []AccelReading{
		{float32(10 * 0.04), float32(-12 * 0.04), float32(-256 * 0.04)},
		{float32(22 * 0.04), float32(-10 * 0.04), float32(-256 * 0.04)},
		{float32(20 * 0.04), float32(-10 * 0.04), float32(-248 * 0.04)},
	})
}

func TestParseBlePayload_IBS03TP(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100D809060A640017040000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03TP")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Temperature", float32(25.20))
	validateFieldFunc(t, got, "TemperatureExt", float32(25.66))
}

func TestParseBlePayload_IBS03R(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC280100AAAA7200000013090000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03R")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
	validateFieldFunc(t, got, "Range", int16(114))
	validateFieldFunc(t, got, "Temperature", nil)
}

func TestParseBlePayload_IBS03P(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC2C0100BF0AD00A0000120A0600")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03P")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3))
	validateFieldFunc(t, got, "Range", nil)
	validateFieldFunc(t, got, "Temperature", float32(27.51))
	validateFieldFunc(t, got, "TemperatureExt", float32(27.68))
	validateFieldFunc(t, got, "Humidity", nil)
}

func TestParseBlePayload_IBS03GP(t *testing.T) {
	payload, _ := hex.DecodeString("0201061BFF0D0085BC3111160082FF9EFE4E001200D2FE10003A005CFFD9C5")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS03GP")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.05))
	validateFieldFunc(t, got, "Accels", []AccelReading{
		{float32(22 * 0.04), float32(-126 * 0.04), float32(-354 * 0.04)},
		{float32(78 * 0.04), float32(18 * 0.04), float32(-302 * 0.04)},
		{float32(16 * 0.04), float32(58 * 0.04), float32(-164 * 0.04)},
	})
	validateFieldFunc(t, got, "Moving", true)
}

func TestParseBlePayload_IBS04(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC3A0101AAAAFFFF000019070000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS04")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.14))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

// test SCAN RESPONSE of iBS04i
func TestParseBlePayload_IBS04i(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC1F0100AAAAFFFF000018030000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS04i")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.87))
	validateFieldFunc(t, got, "ButtonPressed", false)
}

func TestParseBlePayload_IBS05(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC290101AAAAFFFF000030000000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "ButtonPressed", true)
}

func TestParseBlePayload_IBS05T(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC4A0100A10AFFFF000032000000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05T")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "Temperature", float32(27.21))
	validateFieldFunc(t, got, "Humidity", nil)
}

func TestParseBlePayload_IBS05G(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC290102AAAAFFFF000033000000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05G")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.97))
	validateFieldFunc(t, got, "ButtonPressed", false)
	validateFieldFunc(t, got, "Moving", true)
}

func TestParseBlePayload_IBS05CO2(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC270100AAAA6804000034010000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS05CO2")
	validateFieldFunc(t, got, "Temperature", nil)
}

func TestParseBlePayload_IBS06(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF2C0883BC4A0100AAAAFFFF000040110000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS06")
	validateFieldFunc(t, got, "BatteryVoltage", float32(3.3))
}

func TestParseBlePayload_IBS02HM(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0082BC280100AAAAFFFF000004050000")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iBS02HM")
	validateFieldFunc(t, got, "BatteryVoltage", float32(2.96))
}

func TestParseBlePayload_IRS02RG(t *testing.T) {
	payload, _ := hex.DecodeString("02010612FF0D0083BC4D010000002400FCFE22074B58")
	got := ParseBlePayload(payload)
	validateFieldFunc(t, got, "ProductModel", "iRS02RG")
	validateFieldFunc(t, got, "Accel", AccelReading{
		float32(0 * 0.04), float32(36 * 0.04), float32(-260 * 0.04),
	})
}
