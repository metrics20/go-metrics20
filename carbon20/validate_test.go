package carbon20

import (
	"testing"

	"github.com/bmizerany/assert"
)

func TestValidateLegacy(t *testing.T) {
	cases := []struct {
		in    string
		level ValidationLevelLegacy
		valid bool
	}{
		{"foo.bar", Strict, true},
		{"foo.bar", Medium, true},
		{"foo.bar", None, true},
		{"foo..bar", Strict, false},
		{"foo..bar", Medium, true},
		{"foo..bar", None, true},
		{"foo..bar.ba::z", Strict, false},
		{"foo..bar.ba::z", Medium, true},
		{"foo..bar.ba::z", None, true},
		{"foo..bar.b\xbdz", Strict, false},
		{"foo..bar.b\xbdz", Medium, false},
		{"foo..bar.b\xbdz", None, true},
		{"foo..bar.b\x00z", Strict, false},
		{"foo..bar.b\x00z", Medium, false},
		{"foo..bar.b\x00z", None, true},
	}
	for _, c := range cases {
		assert.Equal(t, ValidateKeyLegacy(c.in, c.level) == nil, c.valid)
		assert.Equal(t, ValidateKeyLegacyB([]byte(c.in), c.level) == nil, c.valid)
	}
}

func TestValidateM20(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
	}{
		{"foo.bar.aunit=no.baz", false},
		{"foo.bar.UNIT=no.baz", false},
		{"foo.bar.unita=no.bar", false},
	}
	for _, c := range cases {
		assert.Equal(t, ValidateKeyM20(c.in) == nil, c.valid)
		assert.Equal(t, ValidateKeyM20B([]byte(c.in)) == nil, c.valid)
	}
}
func TestValidateM20NoEquals(t *testing.T) {
	cases := []struct {
		in    string
		valid bool
	}{
		{"foo.bar.mtype_is_count.baz", false},
		{"foo.bar.mtype_is_count", false},
		{"mtype_is_count.foo.bar", false},
	}
	for _, c := range cases {
		assert.Equal(t, ValidateKeyM20NoEquals(c.in) == nil, c.valid)
		assert.Equal(t, ValidateKeyM20NoEqualsB([]byte(c.in)) == nil, c.valid)
	}
}

func BenchmarkValidatePacket(b *testing.B) {
	in := []byte("carbon.agents.foo.cache.overflow 123.456 1234567890")
	for i := 0; i < b.N; i++ {
		_, _, _, err := ValidatePacket(in, None)
		if err != nil {
			panic(err)
		}
	}
}
