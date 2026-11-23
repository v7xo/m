package storage

import (
	"hash/crc32"
	"strings"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

const advisoryLockIDSalt uint = 1486364155

// GenerateAdvisoryLockId inspired by rails migrations, see https://goo.gl/8o9bCT
func GenerateAdvisoryLockId(databaseName string, additionalNames ...string) (string, error) { // nolint: golint
	if len(additionalNames) > 0 {
		databaseName = strings.Join(append(additionalNames, databaseName), "\x00")
	}

	sum := crc32.ChecksumIEEE([]byte(databaseName))

	sum = sum * uint32(advisoryLockIDSalt)

	return stringsext.FormatUint(sum), nil
}
