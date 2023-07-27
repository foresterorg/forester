package model

import (
	"bytes"
	"net"
	"sort"
)

type HwAddrSlice []net.HardwareAddr

func (s HwAddrSlice) Len() int {
	return len(s)
}

func (s HwAddrSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less sorts hardware addresses in a way that longer (EUI-48/64 or Infiniband) comes first and
// addresses of the same length are sorted by alphabetical (hex) order.
func (s HwAddrSlice) Less(i, j int) bool {
	if len(s[i]) > len(s[j]) {
		return true
	}
	for k := len(s[i]) - 1; k >= 0; k-- {
		if s[i][k] > s[j][k] {
			return false
		}
	}
	return true
}

// Unique sorts the slice and then returns a copy dropping all duplicate items.
func (s HwAddrSlice) Unique() HwAddrSlice {
	sort.Sort(s)
	result := make(HwAddrSlice, 0, len(s))
	var prev net.HardwareAddr

	for i := range s {
		if bytes.Equal(s[i], prev) {
			continue
		}
		prev = s[i]
		result = append(result, s[i])
	}
	return result
}
