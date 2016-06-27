package carbon20

import (
	"testing"
)

func TestGetVersionB(t *testing.T) {
	cases := []struct {
		in []byte
		v  metricVersion
	}{
		{
			[]byte("service=carbon.instance=foo.unit=Err.target_type=gauge.type=cache_overflow"),
			M20,
		},
		{
			[]byte("service_is_carbon.instance_is_foo.unit_is_Err.target_type_is_gauge.type_is_cache_overflow"),
			M20NoEquals,
		},
		{
			[]byte("carbon.agents.foo.cache.overflow"),
			Legacy,
		},
		{
			[]byte("foo-bar"),
			Legacy,
		},
	}
	for i, c := range cases {
		v := GetVersionB(c.in)
		if v != c.v {
			t.Fatalf("case %d: expected %s, got %s", i, c.v, v)
		}
	}
}

func BenchmarkGetVersionBM20(b *testing.B) {
	in := []byte("service=carbon.instance=foo.unit=Err.target_type=gauge.type=cache_overflow")
	for i := 0; i < b.N; i++ {
		GetVersionB(in)
	}
}

func BenchmarkGetVersionBM20NoEquals(b *testing.B) {
	in := []byte("service_is_carbon.instance_is_foo.unit_is_Err.target_type_is_gauge.type_is_cache_overflow")
	for i := 0; i < b.N; i++ {
		GetVersionB(in)
	}
}

func BenchmarkGetVersionBLegacy(b *testing.B) {
	in := []byte("carbon.agents.foo.cache.overflow")
	for i := 0; i < b.N; i++ {
		GetVersionB(in)
	}
}

func BenchmarkValidatePacketRet(b *testing.B) {
	in := []byte("carbon.agents.foo.cache.overflow 123.456 1234567890")
	for i := 0; i < b.N; i++ {
		_, _, _, err := ValidatePacketRet(in, None)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkValidatePacket(b *testing.B) {
	in := []byte("carbon.agents.foo.cache.overflow 123.456 1234567890")
	for i := 0; i < b.N; i++ {
		err := ValidatePacket(in, None)
		if err != nil {
			panic(err)
		}
	}
}
