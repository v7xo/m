package syncext

import (
	"hash/maphash"
)

const JobPartitionKeyAny = JobPartitionKey("")

var jobPartitionKeyHashSeed = maphash.MakeSeed()

type JobPartitionKey string

func (v JobPartitionKey) HashUint64() uint64 {
	return maphash.String(jobPartitionKeyHashSeed, string(v))
}

func (v JobPartitionKey) Any() bool { return v == JobPartitionKeyAny }
