package model

import "github.com/google/uuid"

type BtrfsInfoDevice struct {
	BackingDevice BlockDeviceID
	Size          DataSize
	UsedSize      DataSize
	ID            int32
}

type BtrfsInfo struct {
	Label         string
	BackingDevice []BtrfsInfoDevice
	UUID          uuid.UUID
}
