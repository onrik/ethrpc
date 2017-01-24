package ethrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseInt(t *testing.T) {
	i, err := ParseInt("0x143")
	assert.Nil(t, err)
	assert.Equal(t, 323, i)

	i, err = ParseInt("143")
	assert.Nil(t, err)
	assert.Equal(t, 323, i)

	i, err = ParseInt("0xaaa")
	assert.Nil(t, err)
	assert.Equal(t, 2730, i)

	i, err = ParseInt("1*29")
	assert.NotNil(t, err)
	assert.Equal(t, 0, i)
}

func TestParseBigInt(t *testing.T) {
	i, err := ParseBigInt("0xabc")
	assert.Nil(t, err)
	assert.Equal(t, int64(2748), i.Int64())

	i, err = ParseBigInt("$%1")
	assert.NotNil(t, err)
}
