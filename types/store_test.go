package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrefixEndBytes(t *testing.T) {
	assert := assert.New(t)

	var testCases = []struct {
		prefix   []byte
		expected []byte
	}{
		{[]byte{byte(55), byte(255), byte(255), byte(0)}, []byte{byte(55), byte(255), byte(255), byte(1)}},
		{[]byte{byte(55), byte(255), byte(255), byte(15)}, []byte{byte(55), byte(255), byte(255), byte(16)}},
		{[]byte{byte(55), byte(200), byte(255)}, []byte{byte(55), byte(201)}},
		{[]byte{byte(55), byte(255), byte(255)}, []byte{byte(56)}},
		{[]byte{byte(255), byte(255), byte(255)}, nil},
		{nil, nil},
	}

	for _, test := range testCases {
		end := PrefixEndBytes(test.prefix)
		assert.Equal(test.expected, end)
	}
}
