package main

import (
	"net"
	"sync"
	"time"
)

type Ping struct {
	Mutex         *sync.RWMutex
	ResultChannel chan *float64
	TrackingHash  map[uint]uint
	Conn          net.Conn
	PingInfo      PingInfo
}

// make a | b in big endian form
func beCombine(a byte, b byte) uint {
	return uint(a)<<8 | uint(b)
}

// pack the 64-bit epoch time into the last 8 bytes (the id, sequence, and
// message data fields respectively). This is probably not RFC friendly but it
// gives us really accurate ping times.
func packTime(msg *[]byte, the_time int64) {
	for i := 0; i < 8; i++ {
		offset := i + 4
		(*msg)[offset] = byte(the_time >> uint(i*8))
	}
}

// the inverse of packTime() -- returns a nanosecond epoch as int64.
func unpackTime(msg *[]byte) int64 {
	result := int64(0)

	for x := 0; x < 8; x++ {
		offset := x + 4
		result |= (int64((*msg)[offset]) << uint(x*8))
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

// sends a echo request. returns the ID and sequence number in a 2 element
// tuple.
func doPing(conn net.Conn) []uint {
	the_time := time.Now().UnixNano()

	msg := make([]byte, 12) // total echo request is 12 bytes, big endian
	msg[0] = 8              // echo request is ICMP type 8

	packTime(&msg, the_time)
	genChecksum(&msg)

	conn.Write(msg)

	return []uint{beCombine(msg[4], msg[5]), beCombine(msg[6], msg[7])}
}

func (ping *Ping) pingReader() {
	for i := 0; i < ping.PingInfo.Count; i++ {
		var msg []byte
		err_set := false

		for {
			msg = make([]byte, 32)                                     // 32 ==  20 byte icmp header + 12 byte icmp echo reply
			ping.Conn.SetReadDeadline(time.Now().Add(2 * time.Second)) // if a ping doesn't come back in 2 seconds...
			num, err := ping.Conn.Read(msg)

			if err != nil {
				err_set = true
				break
			}

			if num == 32 { // fragmentation has not been an issue.
				// strip the icmp header, although we probably *should* use this for state
				// tracking.
				msg = msg[20:]

				// this mess just ensures this is a ping we sent. see the return values
				// in doPing()
				key := beCombine(msg[4], msg[5])

				// if we have something, nuke it from the tracking map and bail for
				// message unpack
				ping.Mutex.RLock()
				if (ping.TrackingHash)[key] == beCombine(msg[6], msg[7]) {
					ping.Mutex.RUnlock()
					ping.Mutex.Lock()
					delete(ping.TrackingHash, key)
					ping.Mutex.Unlock()
					break
				}
				ping.Mutex.RUnlock()
			}
		}

		if !err_set {
			// nano -> milliseconds
			result := float64(time.Now().UnixNano()-unpackTime(&msg)) / 1000000
			ping.ResultChannel <- &result
		}
	}
}

// sends a single icmp echo request. fills the 'ours' map with the generated id
// and sequence for tracking in pingReader().
func (ping *Ping) sendPing() {
	res := doPing(ping.Conn)
	ping.Mutex.Lock()
	ping.TrackingHash[res[0]] = res[1]
	ping.Mutex.Unlock()
}

func NewPing(conn net.Conn, pi PingInfo) *Ping {
	return &Ping{
		TrackingHash:  map[uint]uint{},
		Mutex:         new(sync.RWMutex),
		ResultChannel: make(chan *float64, pi.Count),
		Conn:          conn,
		PingInfo:      pi,
	}
}

//
// does num pings against conn, waits for any we know of to return.
// will wait up to wait seconds for a response, and sends 1 ping at a minimum
// of interval.
//
// returns a tuple of float64 with at maximum num values: rtt in ms.
//
func (pi PingInfo) pingTimes(conn net.Conn) []float64 {
	results := []float64{}

	ping := NewPing(conn, pi)

	go ping.pingReader()

	for i := 0; i < pi.Count; i++ {
		go ping.sendPing()
		time.Sleep(time.Duration(pi.Interval) * time.Second)
	}

	wait_for := time.After(time.Duration(pi.Wait) * time.Second)

	for i := pi.Count; i > 0; {
		select {
		case <-wait_for:
			return results
		case result := <-ping.ResultChannel:
			i--
			results = append(results, *result)
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	return results
}

// connect to a host and ping it. returns a tuple of float64 which contains the
// ping times. calculating packet loss is someone else's problem, but
// len(retval) and count should help.
func (pi PingInfo) connectAndPing(host string) []float64 {
	conn, err := net.Dial("ip4:icmp", host)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	return pi.pingTimes(conn)
}
