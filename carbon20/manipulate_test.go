package carbon20

import (
	"github.com/bmizerany/assert"
	"strings"
	"testing"
)

var out string

type Case struct {
	in  string
	out string
}

func TestDeriveCount(t *testing.T) {
	// metrics 2.0 cases with equals
	cases := []Case{
		Case{"foo.bar.unit=yes.baz", "foo.bar.unit=yesps.baz"},
		Case{"foo.bar.unit=yes", "foo.bar.unit=yesps"},
		Case{"unit=yes.foo.bar", "unit=yesps.foo.bar"},
		Case{"mtype=count.foo.unit=ok.bar", "mtype=rate.foo.unit=okps.bar"},
	}
	for _, c := range cases {
		assert.Equal(t, DeriveCount(c.in, "prefix.", false), c.out)
	}

	// same but with equals
	for i, c := range cases {
		cases[i] = Case{
			strings.Replace(c.in, "=", "_is_", -1),
			strings.Replace(c.out, "=", "_is_", -1),
		}
	}
	for _, c := range cases {
		assert.Equal(t, DeriveCount(c.in, "prefix.", false), c.out)
	}
}

// only 1 kind of stat is enough, cause they all behave the same
func TestStat(t *testing.T) {
	cases := []Case{
		Case{"foo.bar.unit=yes.baz", "foo.bar.unit=yes.baz.stat=max_90"},
		Case{"foo.bar.unit=yes", "foo.bar.unit=yes.stat=max_90"},
		Case{"unit=yes.foo.bar", "unit=yes.foo.bar.stat=max_90"},
		Case{"mtype=count.foo.unit=ok.bar", "mtype=count.foo.unit=ok.bar.stat=max_90"},
	}
	for _, c := range cases {
		assert.Equal(t, Max(c.in, "prefix.", "90", ""), c.out)
	}
	// same but with equals and no percentile
	for i, c := range cases {
		cases[i] = Case{
			strings.Replace(c.in, "=", "_is_", -1),
			strings.Replace(strings.Replace(c.out, "=", "_is_", -1), "max_90", "max", 1),
		}
	}
	for _, c := range cases {
		assert.Equal(t, Max(c.in, "prefix.", "", ""), c.out)
	}
}
func TestRateCountPckt(t *testing.T) {
	cases := []Case{
		Case{"foo.bar.unit=yes.baz", "foo.bar.unit=Pckt.baz.orig_unit=yes.pckt_type=sent.direction=in"},
		Case{"foo.bar.unit=yes", "foo.bar.unit=Pckt.orig_unit=yes.pckt_type=sent.direction=in"},
		Case{"unit=yes.foo.bar", "unit=Pckt.foo.bar.orig_unit=yes.pckt_type=sent.direction=in"},
		Case{"mtype=count.foo.unit=ok.bar", "mtype=count.foo.unit=Pckt.bar.orig_unit=ok.pckt_type=sent.direction=in"},
	}
	for _, c := range cases {
		assert.Equal(t, CountPckt(c.in, "prefix."), c.out)
		c = Case{
			c.in,
			strings.Replace(strings.Replace(c.out, "unit=Pckt", "unit=Pcktps", -1), "mtype=count", "mtype=rate", -1),
		}
		assert.Equal(t, RatePckt(c.in, "prefix."), c.out)
	}
	for _, c := range cases {
		c = Case{
			strings.Replace(c.in, "=", "_is_", -1),
			strings.Replace(c.out, "=", "_is_", -1),
		}
		assert.Equal(t, CountPckt(c.in, "prefix."), c.out)
		c = Case{
			c.in,
			strings.Replace(strings.Replace(c.out, "unit_is_Pckt", "unit_is_Pcktps", -1), "mtype_is_count", "mtype_is_rate", -1),
		}
		assert.Equal(t, RatePckt(c.in, "prefix."), c.out)
	}
}

func BenchmarkDeriveCountsM20Bare(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out = DeriveCount("foo=bar", "prefix-m1.", false)
	}
}

func BenchmarkDeriveCountsM20Proper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out = DeriveCount("foo=bar.unit=yes.mtype=count", "prefix-m1.", false)
	}
}

func BenchmarkDeriveCountsM20NoEqualsBare(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out = DeriveCount("foo_is_bar", "prefix-m1.", false)
	}
}

func BenchmarkDeriveCountsM20NoEqualsProper(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out = DeriveCount("foo_is_bar.unit_is_yes.mtype_is_count", "prefix-m1.", false)
	}
}
