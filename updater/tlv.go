package updater

import (
	"encoding/binary"
	"fmt"
	"io"
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

// ValueToBool reads the TLV Value field as a boolean
func ValueToBool(tlv *TLV) bool {
	return ValueToInt(tlv) != 0
}

// ValueToInt reads the TLV Value field as an integer
func ValueToInt(tlv *TLV) int {
	return int(binary.LittleEndian.Uint32(tlv.Value))
}

// intToValue returns the byte slice representation of an integer that
// can be used in a TLV Value field
func intToValue(i uint32) []byte {
	a := make([]byte, 4)
	binary.LittleEndian.PutUint32(a, i)
	return a
}

// boolToValue returns the byte slice representationo of a
// boolean. That byte slice can be stored in the Value field of a TLV
// record
func boolToValue(b bool) []byte {
	if b {
		return intToValue(1)
	}
	return intToValue(0)
}

// ValueToLong reads the TLV Value field as an int64 (long)
func ValueToLong(tlv *TLV) int64 {
	return int64(binary.LittleEndian.Uint64(tlv.Value))
}

// ValueToByteSlice just return the TLV Value field (byte slice)
func ValueToByteSlice(tlv *TLV) []byte {
	return tlv.Value
}

// ValueToString interprets the TLV Value field as a string and
// returns it
func ValueToString(tlv *TLV) string {
	return string(tlv.Value)
}

func IntValueToBytes(tlv *TLV) []byte {
	return tlv.Value
}

// readTlv reads in one TLV record for the io.Reader. If we reach the
// end of the file or if we hit the END_UIC TLV record return a nil
// (and no error)
func readTlv(r io.Reader) (*TLV, error) {
	var record TLV

	// get the tag for this TLV
	err := binary.Read(r, binary.BigEndian, &record.Tag)
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	if record.Tag == END_IUC {
		return nil, nil
	}

	record.TagString = IUCTags[record.Tag]

	switch record.Tag {
	// handle a d. string type
	case
		DSTRING_IUC_COMPANY_NAME,
		DSTRING_IUC_PRODUCT_NAME,
		DSTRING_IUC_INSTALLED_VERSION,
		DSTRING_IUC_SERVER_FILE_SITE,
		DSTRING_IUC_WYUPDATE_SERVER_SITE,
		DSTRING_IUC_HEADER_IMAGE_ALIGNMENT,
		DSTRING_IUC_HEADER_TEXT_COLOR,
		DSTRING_IUC_HEADER_FILENAME,
		DSTRING_IUC_SIDE_IMAGE_FILENAME,
		DSTRING_IUC_LANGUAGE_CULTURE,
		DSTRING_IUC_LANGUAGE_FILENAME:
		err = binary.Read(r, binary.LittleEndian, &record.DataLength)
		if err != nil {
			return nil, err
		}
	default:
	}

	err = binary.Read(r, binary.LittleEndian, &record.Length)
	if err != nil {
		return nil, err
	}

	record.Value = make([]byte, record.Length)
	_, err = io.ReadFull(r, record.Value)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

// tlvWriteDstring writes out a string as the d. string part of a TLV
// record with the specified tag
func tlvWriteDstring(w io.Writer, tag uint8, s string) error {
	tlv := TLV{
		Tag:        tag,
		DataLength: uint32(len(s) + 4),
		Length:     uint32(len(s)),
		Value:      []byte(s),
	}
	return writeTlv(w, tlv)
}

// tlvWriteString writes out a string as the d. string part of a TLV
// record with the specified tag
func tlvWriteString(w io.Writer, tag uint8, s string) error {
	tlv := TLV{
		Tag:    tag,
		Length: uint32(len(s)),
		Value:  []byte(s),
	}
	return writeTlv(w, tlv)
}

// tlvWriteInt writes out an int as the Value field on a TLV record
// with the specified tag
func tlvWriteInt(w io.Writer, tag uint8, i int) error {
	tlv := TLV{
		Tag:    tag,
		Length: 4,
		Value:  intToValue(uint32(i)),
	}
	return writeTlv(w, tlv)
}

// tlvWriteBool writes out a bool as the Value field of of a TLV
// record with the specified tag
func tlvWriteBool(w io.Writer, tag uint8, b bool) error {
	tlv := TLV{
		Tag:    tag,
		Length: 4,
		Value:  boolToValue(b),
	}
	return writeTlv(w, tlv)
}

// writeTagAsTlv writes data in the correct representation according
// to the supplied tag.
func writeTagAsTlv(w io.Writer, tag uint8, data interface{}) (err error) {

	switch tag {
	case
		DSTRING_IUC_COMPANY_NAME,
		DSTRING_IUC_PRODUCT_NAME,
		DSTRING_IUC_INSTALLED_VERSION,
		DSTRING_IUC_SERVER_FILE_SITE,
		DSTRING_IUC_WYUPDATE_SERVER_SITE,
		DSTRING_IUC_HEADER_IMAGE_ALIGNMENT,
		DSTRING_IUC_HEADER_TEXT_COLOR,
		DSTRING_IUC_HEADER_FILENAME,
		DSTRING_IUC_SIDE_IMAGE_FILENAME,
		DSTRING_IUC_LANGUAGE_CULTURE,
		DSTRING_IUC_LANGUAGE_FILENAME:

		if v, ok := data.(string); ok {
			err = tlvWriteDstring(w, tag, v)
		} else {
			err = fmt.Errorf("incorrect d. string type while writing TLV record")
		}

	case
		STRING_IUC_GUID,
		STRING_IUC_CUSTOM_TITLE_BAR,
		STRING_IUC_PUBLIC_KEY:
		if v, ok := data.(string); ok {
			err = tlvWriteString(w, tag, v)
		} else {
			err = fmt.Errorf("incorrect string type while writing TLV record")
		}

	case INT_IUC_HEADER_TEXT_INDENT:
		if v, ok := data.(int); ok {
			err = tlvWriteInt(w, tag, v)
		} else {
			err = fmt.Errorf("incorrect int type while writing TLV record")
		}

	case BOOL_IUC_HIDE_HEADER_DIVIDER, BOOL_IUC_CLOSE_WYUPDATE, END_IUC:
		if v, ok := data.(bool); ok {
			err = tlvWriteBool(w, tag, v)
		} else {
			err = fmt.Errorf("incorrect bool type while writing TLV record")
		}

	}
	return
}

// writeTlv writes TLV to a io.Writer
func writeTlv(f io.Writer, tlv TLV) (err error) {
	if tlv.Length == 0 {
		// this tag is not needed
		return nil
	}

	if f == nil {
		return fmt.Errorf("nil io.Writer passed to writeTlv")
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
