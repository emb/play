package object

import (
	"strconv"
	"testing"
)

func TestStringHashKey(t *testing.T) {
	tests := []struct {
		first  Str
		second Str
		want   bool
	}{
		{Str("same"), Str("same"), true},
		{Str("something"), Str("different"), false},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			if (tc.first.HashKey() == tc.second.HashKey()) != tc.want {
				t.Errorf("first.HashKey() %s, second.HashKey() %s. Want equality to be %t",
					tc.first.HashKey(), tc.second.HashKey(),
					tc.want)
			}
		})
	}
}
