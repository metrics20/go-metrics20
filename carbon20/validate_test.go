package carbon20

import "testing"

func BenchmarkValidatePacket(b *testing.B) {
	in := []byte("carbon.agents.foo.cache.overflow 123.456 1234567890")
	for i := 0; i < b.N; i++ {
		_, _, _, err := ValidatePacket(in, None)
		if err != nil {
			panic(err)
		}
	}
}
