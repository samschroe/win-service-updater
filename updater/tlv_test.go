package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLV_WriteTLV(t *testing.T) {
	tlv := TLV{Length: 1}
	err := WriteTLV(nil, tlv)
	assert.NotNil(t, err)
}
