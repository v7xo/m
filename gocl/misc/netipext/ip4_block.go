package netipext

import (
	"cmp"
	"fmt"
	"math"
	"math/bits"
	"net/netip"
	"slices"

	"rn.m.vo7.io/vo7io/m/gocl/misc/stringsext"
)

type IP4Block struct {
	start uint32
	end   uint32
}

func (dst *IP4Block) FromUint32(start, end uint32) error {
	if dst == nil {
		return nil
	}

	if start > end {
		return fmt.Errorf("failed to map '%s' from uint32 range '[%d; %d]'; start can't be greater than end", stringsext.TypeStringFor[IP4Block](), start, end)
	}

	*dst = IP4Block{start: start, end: end}

	return nil
}

func (v IP4Block) Start() netip.Addr {
	return netip.AddrFrom4([4]byte{byte(v.start >> 24), byte(v.start >> 16), byte(v.start >> 8), byte(v.start)}) // nolint:mnd,revive
}

func (v IP4Block) End() netip.Addr {
	return netip.AddrFrom4([4]byte{byte(v.end >> 24), byte(v.end >> 16), byte(v.end >> 8), byte(v.end)}) // nolint:mnd,revive
}

func (v IP4Block) String() string {
	return "(" + v.Start().String() + " - " + v.End().String() + ")"
}

func (v IP4Block) ToPrefixes() []netip.Prefix {
	res := make([]netip.Prefix, 0, 1024)
	blockStart, blockEnd := v.start, v.end

	for blockStart <= blockEnd {
		blockSize := uint64(blockEnd-blockStart) + 1
		blockSizeBits := uint32(bits.Len64(blockSize) - 1)
		freeBits := uint32(bits.TrailingZeros32(blockStart))

		shift := min(blockSizeBits, freeBits)
		cidr := int(32 - shift)
		ip := netip.AddrFrom4([4]byte{byte(blockStart >> 24), byte(blockStart >> 16), byte(blockStart >> 8), byte(blockStart)}) // nolint:mnd,revive

		res = append(res, netip.PrefixFrom(ip, cidr))

		usedBlockSize := uint64(1) << shift

		if uint64(math.MaxUint32-blockStart) < usedBlockSize {
			break
		}

		blockStart += uint32(usedBlockSize)
	}

	return res
}

type IP4BlockList []IP4Block

func (v *IP4BlockList) Compact() {
	if v == nil || len(*v) < 2 {
		return
	}

	cp := *v
	ptr := 0

	slices.SortFunc(cp, func(a, b IP4Block) int { return cmp.Compare(a.start, b.start) })

	for next := 1; next < len(cp); next++ {
		if cp[next].start-1 <= cp[ptr].end {
			cp[ptr].end = cp[next].end

			continue
		}

		ptr++

		cp[ptr] = cp[next]
	}

	*v = cp[:ptr+1]
}
