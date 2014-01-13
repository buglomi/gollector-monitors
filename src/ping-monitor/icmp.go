// ICMP routines for the ping monitor.
package main

import (
	"fmt"
)

// make a | b in big endian form
func beCombine(a byte, b byte) uint32 {
	return uint32(a)<<8 | uint32(b)
}

// pack the 64-bit epoch time into the last 8 bytes reversed (the id, sequence, and
// message data fields respectively). This is probably not RFC friendly but it
// gives us really accurate ping times.
func packTime(msg *[]byte, the_time int64) {
	for i := 0; i < 8; i++ {
		(*msg)[i+4] = byte(the_time >> uint(i*8))
		fmt.Println("in", (*msg)[i+4])
	}
}

// the inverse of packTime() -- returns a nanosecond epoch as int64.
func unpackTime(msg *[]byte) int64 {
	result := int64(0)

	for i := 0; i < 8; i++ {
		fmt.Println("out", (*msg)[i+4])
		result |= (int64((*msg)[i+4]) << uint(i*8))
	}

	return result
}

func genChecksum(msg *[]byte) {
	// I copied this from
	// https://code.google.com/p/go/source/browse/ipv4/mockicmp_test.go?repo=net
	//
	// I think it's pretty inefficient -- there seem to be a lot of
	// back-and-forth surrounding big-endian and little-endian conversions, but
	// my bit math is horrible and I'm in no place to think about this while I'm
	// writing this comment.
	//
	// Need to review when I have some time to spend on it.

	s := uint32(0)

	// sum
	for i := 0; i < len(*msg)-1; i += 2 {
		s += uint32(beCombine((*msg)[i+1], (*msg)[i]))
	}

	// rotate 2 bytes? what is this for? I'm just a caveman, alone and confused.
	s = s>>16 + s&0xffff
	s = s + s>>16

	// checksum is one's complement of the one's complement of the sum
	(*msg)[2] ^= byte(^s)
	(*msg)[3] ^= byte(^s >> 8)
}
