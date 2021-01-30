package minivmm

import (
	"testing"
)

func TestConvertSIPrefixedValue(t *testing.T) {
	testConvertSIPrefixedValue(t, "10G", "mebi", "10240")
	testConvertSIPrefixedValue(t, "1024M", "mebi", "1024")
	testConvertSIPrefixedValue(t, "1024", "mebi", "1024")

	testConvertSIPrefixedValue(t, "10G", "gibi", "10")
	testConvertSIPrefixedValue(t, "1024M", "gibi", "1")
	testConvertSIPrefixedValue(t, "1048576K", "gibi", "1")
	testConvertSIPrefixedValue(t, "10M", "gibi", "0")
	testConvertSIPrefixedValue(t, "10", "gibi", "10")

	testConvertSIPrefixedValue(t, "4K", "", "4096")
	testConvertSIPrefixedValue(t, "1M", "", "1048576")
	testConvertSIPrefixedValue(t, "1G", "", "1073741824")

	var err error
	_, err = ConvertSIPrefixedValue("10H", "")
	if err == nil {
		t.Errorf("expected error but it does not occur")
	}
	_, err = ConvertSIPrefixedValue("10GB", "")
	if err == nil {
		t.Errorf("expected error but it does not occur")
	}
	_, err = ConvertSIPrefixedValue("NOTHING", "")
	if err == nil {
		t.Errorf("expected error but it does not occur")
	}
	_, err = ConvertSIPrefixedValue("1234-NOTHING", "")
	if err == nil {
		t.Errorf("expected error but it does not occur")
	}
}

func testConvertSIPrefixedValue(t *testing.T, value, destUnit, expected string) {
	actual, err := ConvertSIPrefixedValue(value, destUnit)
	if err != nil {
		t.Errorf("convert error: %v", err)
	}
	if actual != expected {
		t.Errorf("unexpected vlaue; expected:%s actual:%s", expected, actual)
	}
}
