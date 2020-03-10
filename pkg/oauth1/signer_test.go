package oauth1

import "testing"

func testHMACSHA1Sign(t *testing.T) {
	s := HMACSHA1{ConsumerSecret: "xxxxxxxxxxxxxx=="}
	digest := s.Sign("yyyyyyyyyyyyyyyy", "text")
	if digest != "5MbWEwPzYkw/d5l8SErbJDpi0R8=" {
		t.Errorf(`"%s" != "5MbWEwPzYkw/d5l8SErbJDpi0R8="`, digest)
	}
}

func TestHMACSHA1(t *testing.T) {
	t.Run("Sign", testHMACSHA1Sign)
}
