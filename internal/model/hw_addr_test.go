package model

import (
	"net"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSortEmpty(t *testing.T) {
	var s HwAddrSlice
	sort.Sort(s)
}

func TestSort(t *testing.T) {
	a, _ := net.ParseMAC("aa:bb:cc:dd:ee:ff")
	s := HwAddrSlice{a}
	sort.Sort(s)
}

func TestSortedTwo(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	s := HwAddrSlice{a, b}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{a, b}, s)
}

func TestSortLowTwo(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	s := HwAddrSlice{b, a}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{a, b}, s)
}

func TestSortHighTwo(t *testing.T) {
	a, _ := net.ParseMAC("0a:00:00:00:00:00")
	b, _ := net.ParseMAC("0b:00:00:00:00:00")
	s := HwAddrSlice{b, a}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{a, b}, s)
}

func TestSortEUI48Unsorted(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:00:00:0b")
	s := HwAddrSlice{a, b}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{b, a}, s)
}

func TestSortEUI48Sorted(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	s := HwAddrSlice{a, b}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{a, b}, s)
}

func TestSortAll(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	c, _ := net.ParseMAC("00:00:00:00:00:0c")
	s := HwAddrSlice{b, c, a}
	sort.Sort(s)
	require.Equal(t, HwAddrSlice{a, b, c}, s)
}

func TestUniqueEmpty(t *testing.T) {
	s := make(HwAddrSlice, 0)
	require.Equal(t, 0, len(s.Unique()))
}

func TestUniqueOne(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0a")
	c, _ := net.ParseMAC("00:00:00:00:00:0a")
	s := HwAddrSlice{a, b, c}.Unique()
	require.Equal(t, HwAddrSlice{a}, s)
}

func TestUniqueThree(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	c, _ := net.ParseMAC("00:00:00:00:00:0c")
	s := HwAddrSlice{a, b, c}.Unique()
	require.Equal(t, HwAddrSlice{a, b, c}, s)
}

func TestUniqueAll(t *testing.T) {
	a, _ := net.ParseMAC("00:00:00:00:00:00:00:0a")
	b, _ := net.ParseMAC("00:00:00:00:00:0b")
	c, _ := net.ParseMAC("00:00:00:00:00:0c")
	s := HwAddrSlice{b, c, a}.Unique()
	require.Equal(t, HwAddrSlice{a, b, c}, s)
}
