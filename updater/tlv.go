package updater

import (
	"encoding/binary"
	"os"
)

// wyUpdate value types
const (
	TLV_BOOL = iota
	TLV_BYTE
	TLV_DSTRING
	TLV_INT
	TLV_LONG
	TLV_STRING
)

// TLV type, length, value (and some helpers)
type TLV struct {
	Tag        uint8  // wyUpdate tag
	TagString  string // to help debug
	Type       int    // wyUpdate value type
	DataLength uint32 // used in wyUpdate d. strings
	Length     uint32
	Value      []byte
}

// wyUpdate types
// int – 32‐bit, little‐endian, signed integer
// long – 64‐bit, little‐endian, signed integer

// d. string
// Int that stores String Length ‘N’ + 4,
// Int that stores String Length ‘N’,
// UTF8 string N bytes long

// string
// Int that stores String Length ‘N’,
// UTF8 string N bytes long

func ValueToBool(tlv *TLV) []byte {
	return tlv.Value
}

func ValueToInt(tlv *TLV) int {
	return int(binary.LittleEndian.Uint32(tlv.Value))
}

func ValueToLong(tlv *TLV) int64 {
	return int64(binary.LittleEndian.Uint64(tlv.Value))
}

func ValueToByteSlice(tlv *TLV) []byte {
	return tlv.Value
}

func ValueToString(tlv *TLV) string {
	return string(tlv.Value)
}

func IntValueToBytes(tlv *TLV) []byte {
	return tlv.Value
}

// WriteTLV writes TLV to open file
func WriteTLV(f *os.File, tlv TLV) (err error) {
	if tlv.Length == 0 {
		// this tag is not needed
		return nil
	}

	// write tag
	err = binary.Write(f, binary.BigEndian, tlv.Tag)
	if nil != err {
		return err
	}

	// write data length (for d. strings)
	if tlv.DataLength > 0 {
		err = binary.Write(f, binary.LittleEndian, tlv.DataLength)
		if nil != err {
			return err
		}
	}

	// write length
	err = binary.Write(f, binary.LittleEndian, tlv.Length)
	if nil != err {
		return err
	}

	// write value
	err = binary.Write(f, binary.BigEndian, tlv.Value)
	if nil != err {
		return err
	}

	return nil
}
