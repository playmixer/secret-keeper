package tools

import (
	"testing"
)

func TestGetMD5Hash(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Hello World", "b10a8db164e0754105b7a99be72e3fe5"},
		{"", "d41d8cd98f00b204e9800998ecf8427e"},
		{"a", "0cc175b9c0f1b6a831c399e269772661"},
	}

	for _, test := range tests {
		actual := GetMD5Hash(test.input)
		if actual != test.expected {
			t.Errorf("Expected %s to be %s, but got %s", test.input, test.expected, actual)
		}
	}
}

func TestRandomString(t *testing.T) {
	tests := []int{5, 10, 15}

	for _, test := range tests {
		actual := RandomString(uint(test))
		if len(actual) != test {
			t.Errorf("Expected %d len, but got %d", test, len(actual))
		}
	}
}
