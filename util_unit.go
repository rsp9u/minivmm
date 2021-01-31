package minivmm

import (
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

var validValue = regexp.MustCompile(`^([0-9]+)([KMGTP]?)(.*)$`)

// ConvertSIPrefixedValue converts a value with or without SI prefixes.
// This function calculates with 2^10N (such as Kibi, Mebi) not 10^3N.
func ConvertSIPrefixedValue(prefixedValue string, destUnit string) (string, error) {
	match := validValue.FindStringSubmatch(prefixedValue)
	if len(match) == 0 || len(match[3]) != 0 {
		return "", errors.New("Invalid SI prefixed value")
	}
	value := match[1]
	prefix := match[2]

	shift := 0
	switch destUnit {
	case "kibi":
		shift -= 10
	case "mebi":
		shift -= 20
	case "gibi":
		shift -= 30
	case "tebi":
		shift -= 40
	case "pebi":
		shift -= 50
	}
	switch prefix {
	case "K":
		shift += 10
	case "M":
		shift += 20
	case "G":
		shift += 30
	case "T":
		shift += 40
	case "P":
		shift += 50
	case "":
		shift = 0
	}

	valueInt, err := strconv.Atoi(value)
	if err != nil {
		return "", err
	}

	if shift < 0 {
		return strconv.Itoa(int(valueInt >> uint(-shift))), nil
	} else {
		return strconv.Itoa(int(valueInt << uint(shift))), nil
	}
}
