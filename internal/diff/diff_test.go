package diff

import (
	"reflect"
	"testing"
)

func TestStrings(t *testing.T) {
	tests := []test{
		{A: []string{"hello", "world"}, B: []string{"foo", "world"}, Expected: []string{"hello", "foo"}},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			v := Strings(test.A, test.B)
			if !reflect.DeepEqual(v, test.Expected) {
				t.Errorf("got %+v, expected %+v", v, test.Expected)
			}
		})
	}
}

type test struct {
	A        []string
	B        []string
	Expected []string
}
