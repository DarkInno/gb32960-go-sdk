package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/darkinno/gb32960-go-sdk/codec"
	"github.com/darkinno/gb32960-go-sdk/constant"
)

var (
	conns   = flag.Int("c", 500, "total connections")
	workers = flag.Int("w", 100, "max concurrent workers")
	addr    = flag.String("addr", "127.0.0.1:32960", "server address")
)

type stats struct {
	conned       atomic.Int64
	failed       atomic.Int64
	authed       atomic.Int64
	realdata     atomic.Int64
	heartbeats   atomic.Int64
	bytesSent    atomic.Int64
	bytesRecv    atomic.Int64
	totalLatency atomic.Int64
	latencyCount atomic.Int64
}

func main() {
	flag.Parse()

	log.Printf("stress test: %d conns, %d workers -> %s", *conns, *workers, *addr)
	s := &stats{}
	start := time.Now()

	sem := make(chan struct{}, *workers)
	var wg sync.WaitGroup
	var done atomic.Int64

	stopMonitor := make(chan struct{})
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stopMonitor:
				return
			case <-ticker.C:
				d := done.Load()
				elapsed := time.Since(start)
				if elapsed > 30*time.Second {
					return
				}
				fmt.Printf("[%5.1fs] done=%d/%d auth=%d data=%d hb=%d fail=%d\n",
					elapsed.Seconds(), d, *conns,
					s.authed.Load(), s.realdata.Load(), s.heartbeats.Load(), s.failed.Load())
			}
		}
	}()

	for i := 0; i < *conns; i++ {
		sem <- struct{}{}
		wg.Add(1)
		go func(id int) {
			defer func() {
				wg.Done()
				done.Add(1)
				<-sem
			}()
			runClient(id, s)
		}(i)
	}

	wg.Wait()
	close(stopMonitor)

	elapsed := time.Since(start)
	fmt.Println()
	fmt.Println("=======================================")
	fmt.Println("  Stress Test Report")
	fmt.Println("=======================================")
	fmt.Printf("  target       : %s\n", *addr)
	fmt.Printf("  total conns  : %d\n", *conns)
	fmt.Printf("  workers      : %d\n", *workers)
	fmt.Printf("  duration     : %v\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  throughput   : %.0f conns/s\n", float64(*conns)/elapsed.Seconds())
	fmt.Println("---------------------------------------")
	fmt.Printf("  connected    : %d\n", s.conned.Load())
	fmt.Printf("  failed       : %d\n", s.failed.Load())
	fmt.Printf("  authenticated: %d\n", s.authed.Load())
	fmt.Printf("  realtime     : %d pkts\n", s.realdata.Load())
	fmt.Printf("  heartbeats   : %d\n", s.heartbeats.Load())
	fmt.Printf("  sent         : %d (%.1f MB)\n", s.bytesSent.Load(), float64(s.bytesSent.Load())/1e6)
	fmt.Printf("  recv         : %d (%.1f MB)\n", s.bytesRecv.Load(), float64(s.bytesRecv.Load())/1e6)
	if c := s.latencyCount.Load(); c > 0 {
		avg := time.Duration(s.totalLatency.Load() / c)
		fmt.Printf("  avg login lat: %v\n", avg.Round(time.Microsecond))
	}
	rate := float64(s.failed.Load()) / float64(*conns) * 100
	fmt.Printf("  fail rate    : %.1f%%\n", rate)
	fmt.Println("=======================================")
}

func runClient(id int, s *stats) {
	vin := fmt.Sprintf("HC%013d", id)

	conn, err := net.DialTimeout("tcp", *addr, 5*time.Second)
	if err != nil {
		s.failed.Add(1)
		return
	}
	s.conned.Add(1)

	loginStart := time.Now()

	loginData := []byte{
		0x26, 0x05, 0x0E, 0x0E, 0x30, 0x00,
		byte(id >> 8), byte(id),
		0x10,
		'S', 'I', 'M', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '0', '1',
		0x00,
	}
	if !sendAndRecv(conn, constant.CmdLogin, vin, loginData, s) {
		conn.Close()
		s.failed.Add(1)
		return
	}
	s.totalLatency.Add(int64(time.Since(loginStart)))
	s.latencyCount.Add(1)
	s.authed.Add(1)

	for i := 0; i < 2; i++ {
		rt := buildRealtime(id, i)
		if !sendOnly(conn, constant.CmdRealtime, vin, rt, s) {
			conn.Close()
			return
		}
		s.realdata.Add(1)
	}

	if !sendOnly(conn, constant.CmdHeartbeat, vin, nil, s) {
		conn.Close()
		return
	}
	s.heartbeats.Add(1)

	logoutData := []byte{0x26, 0x05, 0x0E, 0x0E, 0x35, 0x00, 0x00, 0x03}
	sendOnly(conn, constant.CmdLogout, vin, logoutData, s)
	conn.Close()
}

func sendAndRecv(conn net.Conn, cmd byte, vin string, data []byte, s *stats) bool {
	pkt := buildPacket(cmd, vin, data)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Write(pkt)
	if err != nil {
		return false
	}
	s.bytesSent.Add(int64(n))

	buf := make([]byte, 256)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err = conn.Read(buf)
	if err != nil {
		return n > 0
	}
	s.bytesRecv.Add(int64(n))
	return true
}

func sendOnly(conn net.Conn, cmd byte, vin string, data []byte, s *stats) bool {
	pkt := buildPacket(cmd, vin, data)
	conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Write(pkt)
	if err != nil {
		return false
	}
	s.bytesSent.Add(int64(n))
	return true
}

func buildRealtime(id, seq int) []byte {
	data := make([]byte, 0, 42)
	data = append(data, 0x26, 0x05, 0x0E, 0x0E, 0x30, byte(seq*10))
	data = append(data, codec.FieldVehicle)
	v := make([]byte, 20)
	v[0], v[1], v[2] = 0x01, 0x01, 0x01
	binary.BigEndian.PutUint16(v[3:5], uint16(400+seq*50))
	binary.BigEndian.PutUint32(v[5:9], uint32(123450+seq*100))
	binary.BigEndian.PutUint16(v[9:11], 380)
	binary.BigEndian.PutUint16(v[11:13], 200)
	v[13] = 85
	data = append(data, v...)
	data = append(data, codec.FieldPosition)
	p := make([]byte, 8)
	binary.BigEndian.PutUint32(p[0:4], uint32(121473708))
	binary.BigEndian.PutUint32(p[4:8], uint32(31235538))
	data = append(data, p...)
	return data
}

func buildPacket(command byte, vin string, data []byte) []byte {
	vinBytes := make([]byte, constant.VINLength)
	copy(vinBytes, []byte(vin))
	totalLen := constant.HeaderSize + len(data) + 1
	pkt := make([]byte, totalLen)
	pos := 0
	pkt[pos] = constant.StartMarker1
	pos++
	pkt[pos] = constant.StartMarker2
	pos++
	pkt[pos] = command
	pos++
	pkt[pos] = 0x01
	pos++
	copy(pkt[pos:pos+constant.VINLength], vinBytes)
	pos += constant.VINLength
	pkt[pos] = constant.EncNone
	pos++
	binary.BigEndian.PutUint16(pkt[pos:pos+2], uint16(len(data)))
	pos += 2
	copy(pkt[pos:pos+len(data)], data)
	pos += len(data)
	var bcc byte
	for i := 2; i < pos; i++ {
		bcc ^= pkt[i]
	}
	pkt[pos] = bcc
	return pkt
}
