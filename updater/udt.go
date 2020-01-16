// Parser for updtdetails.udt (update details)
// File ID: IUUDFV2
// Filename: updtdetails.udt

package updater

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// UDT tags
const (
	INT_UDT_NUMBER_OF_REGISTRY_CHANGES           = 0x20
	INT_UDT_NUMBER_OF_FILE_INFOS                 = 0x21 // (precedes file info list)
	UDT_BEGINNING_OF_FILE_INFORMATION_IDENTIFIER = 0x8B
	UDT_RELATIVE_FILE_PATH_DSTRING               = 0x40
	UDT_DELTA_PATCH_RELATIVE_PATH_DSTRING        = 0x47
	UDT_NEW_FILES_ADLER32_CHECKSUM_LONG          = 0x48
	UDT_END_OF_FILE_INFO_IDENTIFIER              = 0x9B
	STRING_UDT_SERVICE_TO_STOP_BEFORE_UPDATE     = 0x32
	STRING_UDT_SERVICE_TO_START_AFTER_UPDATE     = 0x33
	END_UDT                                      = 0xFF
)

// UDT tag to string mapping
var UDTTags = map[uint8]string{
	INT_UDT_NUMBER_OF_REGISTRY_CHANGES:           "INT_UDT_NUMBER_OF_REGISTRY_CHANGES",
	INT_UDT_NUMBER_OF_FILE_INFOS:                 "INT_UDT_NUMBER_OF_FILE_INFOS", // (precedes file info list)
	UDT_BEGINNING_OF_FILE_INFORMATION_IDENTIFIER: "UDT_BEGINNING_OF_FILE_INFORMATION_IDENTIFIER",
	UDT_RELATIVE_FILE_PATH_DSTRING:               "UDT_RELATIVE_FILE_PATH_DSTRING",
	UDT_DELTA_PATCH_RELATIVE_PATH_DSTRING:        "UDT_DELTA_PATCH_RELATIVE_PATH_DSTRING",
	UDT_NEW_FILES_ADLER32_CHECKSUM_LONG:          "UDT_NEW_FILES_ADLER32_CHECKSUM_LONG",
	UDT_END_OF_FILE_INFO_IDENTIFIER:              "UDT_END_OF_FILE_INFO_IDENTIFIER",
	STRING_UDT_SERVICE_TO_STOP_BEFORE_UPDATE:     "STRING_UDT_SERVICE_TO_STOP_BEFORE_UPDATE",
	STRING_UDT_SERVICE_TO_START_AFTER_UPDATE:     "STRING_UDT_SERVICE_TO_START_AFTER_UPDATE",
	END_UDT: "END_UDT",
}

type ConfigUDT struct {
	ServiceToStopBeforeUpdate []TLV
	ServiceToStartAfterUpdate []TLV
	NumberOfFileInfos         TLV
	NumberOfRegistryChanges   TLV
}

// ReadUDTTLV reads a single TLV and returns it
func ReadUDTTLV(r io.Reader) (*TLV, error) {
	var record TLV

	err := binary.Read(r, binary.BigEndian, &record.Tag)
	if err == io.EOF {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	record.TagString = UDTTags[record.Tag]

	if record.Tag == END_UDT {
		return nil, nil
	}

	if record.Tag == UDT_BEGINNING_OF_FILE_INFORMATION_IDENTIFIER || record.Tag == UDT_END_OF_FILE_INFO_IDENTIFIER {
		return &record, nil
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

// ParseUDT parses a updtdetails.udt file
func ParseUDT(path string) (ConfigUDT, error) {
	var udt ConfigUDT

	f, err := os.Open(path)
	if nil != err {
		return udt, err
	}
	defer f.Close()

	// read HEADER
	header := make([]byte, 7)
	f.Read(header)

	if string(header) != UPDTDETAILS_HEADER {
		err := fmt.Errorf("invalid update details file")
		return udt, err
	}

	for {
		tlv, err := ReadUDTTLV(f)
		if nil != err {
			return udt, err
		}
		if tlv == nil && nil == err {
			break
		}

		switch tlv.Tag {
		case STRING_UDT_SERVICE_TO_STOP_BEFORE_UPDATE:
			udt.ServiceToStopBeforeUpdate = append(udt.ServiceToStopBeforeUpdate, *tlv)
		case STRING_UDT_SERVICE_TO_START_AFTER_UPDATE:
			udt.ServiceToStartAfterUpdate = append(udt.ServiceToStartAfterUpdate, *tlv)
		case INT_UDT_NUMBER_OF_REGISTRY_CHANGES:
			udt.NumberOfRegistryChanges = *tlv
		case INT_UDT_NUMBER_OF_FILE_INFOS:
			udt.NumberOfFileInfos = *tlv
		case UDT_BEGINNING_OF_FILE_INFORMATION_IDENTIFIER:
			continue
		case UDT_RELATIVE_FILE_PATH_DSTRING:
			continue
		case UDT_DELTA_PATCH_RELATIVE_PATH_DSTRING:
			continue
		case UDT_NEW_FILES_ADLER32_CHECKSUM_LONG:
			continue
		case UDT_END_OF_FILE_INFO_IDENTIFIER:
			continue
		default:
			err := fmt.Errorf("udt tag %x not implemented", tlv.Tag)
			return udt, err
		}
	}

	return udt, err
}

// WriteUDT writes a UDT file
// Not all wyUpdate UDT options are implemented
// This is currently only used in testing
func WriteUDT(udt ConfigUDT, path string) error {
	f, err := os.Create(path)
	if nil != err {
		return err
	}
	defer f.Close()

	// write HEADER
	f.Write([]byte(UPDTDETAILS_HEADER))

	// INT_UDT_NUMBER_OF_REGISTRY_CHANGES
	err = WriteTLV(f, udt.NumberOfRegistryChanges)
	if nil != err {
		return err
	}

	// INT_UDT_NUMBER_OF_FILE_INFOS
	err = WriteTLV(f, udt.NumberOfFileInfos)
	if nil != err {
		return err
	}

	// STRING_UDT_SERVICE_TO_STOP_BEFORE_UPDATE
	for _, s := range udt.ServiceToStopBeforeUpdate {
		err := WriteTLV(f, s)
		if nil != err {
			return err
		}
	}

	// STRING_UDT_SERVICE_TO_START_AFTER_UPDATE
	for _, s := range udt.ServiceToStartAfterUpdate {
		err := WriteTLV(f, s)
		if nil != err {
			return err
		}
	}

	err = binary.Write(f, binary.BigEndian, byte(END_UDT))
	if nil != err {
		return err
	}

	return nil
}
