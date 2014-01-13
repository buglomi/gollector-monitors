// ICMP routines for the ping monitor.
package main

import (
	"fmt"
	metrics "github.com/rcrowley/go-metrics"
	"math"
	"net"
	"sync"
	"time"
)

type PingInfo struct {
	Count    int
	Wait     float64
	Interval float64
	Repeat   int
	Hosts    []string
}

type Ping struct {
	Mutex         *sync.RWMutex
	ResultChannel chan int64
	TrackingHash  map[uint32]uint32
	Conn          net.Conn
	PingInfo      PingInfo
}

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

// sends a echo request. returns the ID and sequence number in a 2 element
// tuple.
func doPing(conn net.Conn, send_chan chan []uint32) {
	the_time := time.Now().UnixNano()

	msg := make([]byte, 12) // total echo request is 12 bytes, big endian
	msg[0] = 8              // echo request is ICMP type 8

	packTime(&msg, the_time)
	send_chan <- []uint32{beCombine(msg[4], msg[5]), beCombine(msg[6], msg[7])}
	genChecksum(&msg)

	conn.Write(msg)
}

func (ping *Ping) pingReader() {
	for i := 0; i < ping.PingInfo.Count; i++ {
		var msg []byte

		for {
			msg = make([]byte, 32)                                                                     // 32 ==  20 byte icmp header + 12 byte icmp echo reply
			ping.Conn.SetReadDeadline(time.Now().Add(time.Duration(ping.PingInfo.Wait) * time.Second)) // if a ping doesn't come back in 2 seconds...
			num, err := ping.Conn.Read(msg)

			if err != nil {
				continue
			}

			if num == 32 { // fragmentation has not been an issue.

				// strip the icmp header, although we probably *should* use this for state
				// tracking.
				msg = msg[20:]

				if msg[0] != 0 {
					continue
				}

				// this mess just ensures this is a ping we sent. see the return values
				// in doPing()
				key := beCombine(msg[4], msg[5])

				// if we have something, nuke it from the tracking map and bail for
				// message unpack
				ping.Mutex.RLock()
				if (ping.TrackingHash)[key] == beCombine(msg[6], msg[7]) {
					fmt.Println(ping.TrackingHash, key, beCombine(msg[6], msg[7]), msg, num, err)
					ping.Mutex.RUnlock()
					ping.Mutex.Lock()
					delete(ping.TrackingHash, key)
					ping.Mutex.Unlock()
					fmt.Println(time.Now().UnixNano(), unpackTime(&msg), (time.Now().UnixNano() - unpackTime(&msg)))
					ping.ResultChannel <- (time.Now().UnixNano() - unpackTime(&msg))
					break
				}
				ping.Mutex.RUnlock()
			}
		}
	}
}

// sends a single icmp echo request. fills the 'ours' map with the generated id
// and sequence for tracking in pingReader().
func (ping *Ping) sendPing() {
	send_chan := make(chan []uint32, 100)
	go doPing(ping.Conn, send_chan)
	for {
		select {
		case res := <-send_chan:
			ping.Mutex.Lock()
			ping.TrackingHash[res[0]] = res[1]
			ping.Mutex.Unlock()
			fmt.Println(ping.TrackingHash)
			return
		}
	}
}

func NewPing(conn net.Conn, pi PingInfo) *Ping {
	return &Ping{
		TrackingHash:  map[uint32]uint32{},
		Mutex:         new(sync.RWMutex),
		ResultChannel: make(chan int64, pi.Count),
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
func (pi PingInfo) pingTimes(conn net.Conn, registry *metrics.Registry) {
	count := 0

	ping := NewPing(conn, pi)

	go ping.pingReader()
	wait_for := time.After(time.Duration(pi.Wait) * time.Second)

	for i := 0; i < pi.Count; i++ {
		ping.sendPing()
		time.Sleep(time.Duration(ping.PingInfo.Interval) * time.Second)
	}

	timeout := false

	for i := pi.Count; i != 0 && !timeout; {
		select {
		case result := <-ping.ResultChannel:
			fmt.Println(result)
			i--
			count++
			metrics.GetOrRegisterHistogram(
				"ns",
				*registry,
				metrics.NewUniformSample(pi.Count),
			).Update(result)
		case <-wait_for:
			timeout = true
			break
		default:
			time.Sleep(1 * time.Millisecond)
		}
	}

	update := int64(math.Floor(float64(count) / float64(pi.Count) * 100))

	metrics.GetOrRegisterHistogram(
		"success",
		*registry,
		metrics.NewUniformSample(10),
	).Update(update)
}

// connect to a host and ping it. returns a tuple of float64 which contains the
// ping times. calculating packet loss is someone else's problem, but
// len(retval) and count should help.
func (pi PingInfo) connectAndPing(host string, registry *metrics.Registry) {
	conn, err := net.Dial("ip4:icmp", host)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	pi.pingTimes(conn, registry)
}
