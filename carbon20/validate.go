package carbon20

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var errTooManyEquals = errors.New("more than 1 equals")
var errKeyOrValEmpty = errors.New("tag_k and tag_v must be non-empty strings")
var errWrongNumFields = errors.New("packet must consist of 3 fields")
var errValNotNumber = errors.New("value field is not a float or int")
var errTsNotTs = errors.New("timestamp field is not a unix timestamp")
var errEmptyNode = errors.New("empty node")
var errMixEqualsTypes = errors.New("both = and _is_")
var errNoUnit = errors.New("no unit tag")
var errNoTargetType = errors.New("no target_type tag")
var errNotEnoughTags = errors.New("must have at least 1 tag beyond unit and target_type")

var errFmtNullAt = "null byte at position %d"
var errFmtIllegalChar = "illegal char %q"
var errFmtNonAsciiChar = "non-ASCII char %q"

// LegacyMetricValidation indicates the level of validation to undertake for legacy metrics
//go:generate stringer -type=LegacyMetricValidation
type LegacyMetricValidation int

const (
	Strict LegacyMetricValidation = iota // Sensible character validation and no consecutive dots
	Medium                               // Ensure characters are 8-bit clean and not NULL
	None                                 // No validation
)

// ValidateSensibleChars checks that the metric id only contains characters that
// are commonly understood to be sensible and useful.  Because Graphite will do
// the weirdest things with all kinds of special characters.
func ValidateSensibleChars(metric_id string) error {
	for _, ch := range metric_id {
		if !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') && ch != '_' && ch != '-' && ch != '.' {
			return fmt.Errorf(errFmtIllegalChar, ch)
		}
	}
	return nil
}

// ValidateSensibleCharsB is like ValidateSensibleChars but for byte array inputs.
func ValidateSensibleCharsB(metric_id []byte) error {
	for _, ch := range metric_id {
		if !(ch >= 'a' && ch <= 'z') && !(ch >= 'A' && ch <= 'Z') && !(ch >= '0' && ch <= '9') && ch != '_' && ch != '-' && ch != '.' {
			return fmt.Errorf(errFmtIllegalChar, ch)
		}
	}
	return nil
}

// validateNotNullAsciiChars returns true if all bytes in metric_id are 8-bit
// clean and no byte is a NULL byte. Otherwise, it returns false.
func validateNotNullAsciiChars(metric_id []byte) error {
	for i, ch := range metric_id {
		if ch == 0 {
			return fmt.Errorf(errFmtNullAt, i)
		}
		if ch&0x80 != 0 {
			return fmt.Errorf(errFmtNonAsciiChar, ch)
		}
	}
	return nil
}

// InitialValidation checks the basic form of metric keys
func InitialValidation(metric_id string, version metricVersion) error {
	if version == Legacy {
		// if the metric contains no = or _is_, in theory we don't really care what it does contain.  it can be whatever.
		// in practice, graphite alters (removes a dot) the metric id when this happens:
		if strings.Contains(metric_id, "..") {
			return errEmptyNode
		}
		return ValidateSensibleChars(metric_id)
	}
	if version == M20 {
		if strings.Contains(metric_id, "_is_") {
			return errMixEqualsTypes
		}
		if !strings.HasPrefix(metric_id, "unit=") && !strings.Contains(metric_id, ".unit=") {
			return errNoUnit
		}
		if !strings.HasPrefix(metric_id, "target_type=") && !strings.Contains(metric_id, ".target_type=") {
			return errNoTargetType
		}
	} else { //version == M20NoEquals
		if strings.Contains(metric_id, "=") {
			return errMixEqualsTypes
		}
		if !strings.HasPrefix(metric_id, "unit_is_") && !strings.Contains(metric_id, ".unit_is_") {
			return errNoUnit
		}
		if !strings.HasPrefix(metric_id, "target_type_is_") && !strings.Contains(metric_id, ".target_type_is_") {
			return errNoTargetType
		}
	}
	if strings.Count(metric_id, ".") < 2 {
		return errNotEnoughTags
	}
	return nil
}

// optimization so compiler doesn't initialize and allocate new variables every time we use this.
// shouldn't be needed for the strings above because they are immutable, I'm assuming the compiler optimizes for that
var (
	doubleDot    = []byte("..")
	m20Is        = []byte("_is_")
	m20UnitPre   = []byte("unit=")
	m20UnitMid   = []byte(".unit=")
	m20TTPre     = []byte("target_type=")
	m20TTMid     = []byte(".target_type=")
	m20NEIS      = []byte("=")
	m20NEUnitPre = []byte("unit_is_")
	m20NEUnitMid = []byte(".unit_is_")
	m20NETTPre   = []byte("target_type_is_")
	m20NETTMid   = []byte(".target_type_is_")
	dot          = []byte(".")
)

// InitialValidationB is like InitialValidation but for byte array inputs.
func InitialValidationB(metric_id []byte, version metricVersion, legacyValidation LegacyMetricValidation) error {
	if version == Legacy {
		if legacyValidation == Strict {
			if bytes.Contains(metric_id, doubleDot) {
				return errEmptyNode
			}
			return ValidateSensibleCharsB(metric_id)
		} else if legacyValidation == Medium {
			return validateNotNullAsciiChars(metric_id)
		}
	} else {
		if version == M20 {
			if bytes.Contains(metric_id, m20Is) {
				return errMixEqualsTypes
			}
			if !bytes.HasPrefix(metric_id, m20UnitPre) && !bytes.Contains(metric_id, m20UnitMid) {
				return errNoUnit
			}
			if !bytes.HasPrefix(metric_id, m20TTPre) && !bytes.Contains(metric_id, m20TTMid) {
				return errNoTargetType
			}
		} else { //version == M20NoEquals
			if bytes.Contains(metric_id, m20NEIS) {
				return errMixEqualsTypes
			}
			if !bytes.HasPrefix(metric_id, m20NEUnitPre) && !bytes.Contains(metric_id, m20NEUnitMid) {
				return errNoUnit
			}
			if !bytes.HasPrefix(metric_id, m20NETTPre) && !bytes.Contains(metric_id, m20NETTMid) {
				return errNoTargetType
			}
		}
		if bytes.Count(metric_id, dot) < 2 {
			return errNotEnoughTags
		}
	}
	return nil
}

var space = []byte(" ")
var empty = []byte("")

// ValidatePacket validates a carbon message and returns useful pieces of it
func ValidatePacket(buf []byte, legacyValidation LegacyMetricValidation) ([]byte, float64, uint32, error) {
	fields := bytes.Fields(buf)
	if len(fields) != 3 {
		return empty, 0, 0, errWrongNumFields
	}

	version := GetVersionB(fields[0])
	err := InitialValidationB(fields[0], version, legacyValidation)
	if err != nil {
		return empty, 0, 0, err
	}

	val, err := strconv.ParseFloat(string(fields[1]), 32)
	if err != nil {
		return empty, 0, 0, errValNotNumber
	}

	ts, err := strconv.ParseUint(string(fields[2]), 10, 0)
	if err != nil {
		return empty, 0, 0, errTsNotTs
	}

	return fields[0], val, uint32(ts), nil
}