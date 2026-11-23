package model

import (
	"errors"
	"fmt"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext/placeholder"
)

type CryptDeviceStatus struct {
	device        BlockDeviceID
	BackingDevice BlockDeviceID
	Cipher        CryptCipherParams
	KeyLocation   string
	Mode          ModeFlag
	Flags         CryptDeviceFlag
	Type          CryptDeviceTypeCode
	KeySizeBits   KeySizeBits
}

//nolint:gocognit,maintidx,gofumpt
func (v *CryptDeviceStatus) FromCryptsetupStatusOutput(val [][]byte, id BlockDeviceID) (err error) {
	if v == nil {
		return
	}

	defer func() {
		if err != nil {
			err = fmt.Errorf("failed to map crypt device status model from 'cryptsetup' command output; err: %w", err)
		}
	}()

	if id == nil {
		return errors.New("no device ID argument specified")
	}

	fieldNameBuf := make([]byte, 0, 96) // nolint
	fieldValBuf := make([]byte, 0, 256) // nolint
	cp := CryptDeviceStatus{device: id}

	const (
		cipherFieldName      = "cipher"
		deviceFieldName      = "device"
		flagsFieldName       = "flags"
		keyLocationFieldName = "key location"
		keySizeFieldName     = "keysize"
		modeFieldName        = "mode"
		typeFieldName        = "type"

		parseContextFieldName byte = iota - 1
		parseContextFieldVal
	)

	for i := range val {
		fieldNameBuf = fieldNameBuf[:0]
		fieldValBuf = fieldValBuf[:0]

		currentParseContext := parseContextFieldName

		for _, sym := range val[i] {
			switch sym {
			case '\t', ' ':
				if currentParseContext == parseContextFieldName {
					if len(fieldNameBuf) == 0 {
						continue
					}

					fieldNameBuf = append(fieldNameBuf, sym)
				} else if len(fieldValBuf) == 0 {
					continue
				} else {
					fieldValBuf = append(fieldValBuf, sym)
				}
			case ':':
				if currentParseContext == parseContextFieldName {
					currentParseContext = parseContextFieldVal

					continue
				}

				fieldValBuf = append(fieldValBuf, sym)
			default:
				if currentParseContext == parseContextFieldName {
					fieldNameBuf = append(fieldNameBuf, sym)

					continue
				}

				fieldValBuf = append(fieldValBuf, sym)
			}
		}

		fieldName := string(fieldNameBuf)

		switch fieldName {
		case typeFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies crypt device type")
			}

			if err = cp.Type.FromCryptsetupStatus(string(fieldValBuf)); err != nil {
				return
			}
		case deviceFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies backing device")
			}

			bPath := string(fieldValBuf)
			bType := BlockDeviceTypeCodeUnspecified
			bBacking := BlockDeviceID(nil)

			if id.BackingDevice() != nil {
				switch {
				case bPath != id.BackingDevice().Path():
					return fmt.Errorf("unexpected mismatch of parsed backing device '%s' and explicitly specified by device ID argument '%s'",
						bPath,
						id.BackingDevice().Path())
				}

				bType = id.BackingDevice().Type()
				bBacking = id.BackingDevice().BackingDevice()
			}

			cp.BackingDevice = NewBlockDeviceID(bPath, bType, BlockDeviceIDOption().Backing(bBacking))
		case cipherFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies cipher")
			}

			if cp.Cipher, err = DecodeCryptCipherParams(string(fieldValBuf)); err != nil {
				return
			}
		case modeFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies mode")
			}

			if err = cp.Mode.FromCryptsetupOutput(string(fieldValBuf)); err != nil {
				return
			}
		case keyLocationFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies key location")
			}

			cp.KeyLocation = string(fieldValBuf)
		case keySizeFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies key size")
			}

			if err = cp.KeySizeBits.FromCryptsetupOutput(string(fieldValBuf)); err != nil {
				return
			}
		case flagsFieldName:
			if len(fieldValBuf) == 0 {
				return errors.New("output not specifies key size")
			}

			if err = cp.Flags.FromCryptsetupOutput(string(fieldValBuf)); err != nil {
				return
			}
		}
	}

	*v = cp

	return
}

func (v CryptDeviceStatus) Device() BlockDeviceID { return v.device }

func (v CryptDeviceStatus) String() string {
	sb := strings.Builder{}

	sb.WriteByte('(')

	sb.WriteString("device: ")

	if v.device == nil {
		sb.WriteString(placeholder.Nil)
	} else {
		sb.WriteString(v.device.String())
	}

	sb.WriteString(", backing: ")

	if v.BackingDevice == nil {
		sb.WriteString(placeholder.Nil)
	} else {
		sb.WriteString(v.BackingDevice.String())
	}

	sb.WriteByte(')')

	return sb.String()
}
