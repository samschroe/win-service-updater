package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTLV_WriteTLV(t *testing.T) {
	tlv := TLV{Length: 1}
	err := writeTlv(nil, tlv)
	assert.NotNil(t, err)
}
