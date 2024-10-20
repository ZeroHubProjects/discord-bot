package ss13

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

func SendRequest(serverAddress string, request []byte) ([]byte, error) {
	if request[0] != '?' {
		request = append([]byte{'?'}, request...)
	}

	// Prepare a packet in BYOND-specific format for interaction with world/Topic()
	query := []byte{0x00, 0x83}
	packetLength := make([]byte, 2)
	binary.BigEndian.PutUint16(packetLength, uint16(len(request)+6))
	query = append(query, packetLength...)
	query = append(query, []byte{0x00, 0x00, 0x00, 0x00, 0x00}...)
	query = append(query, request...)
	query = append(query, 0x00)

	// Establish a tcp connection
	dialer := new(net.Dialer)
	conn, err := dialer.Dial("tcp", serverAddress)
	if err != nil {
		return nil, err
	}

	// Send the packet
	_, err = conn.Write(query)
	if err != nil {
		return nil, err
	}

	// Receive response
	var data []byte
	buffer := make([]byte, 1024)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			break
		}
		data = append(data, buffer[:n]...)
		if n < 1024 {
			break
		}
	}

	return decodePacket(data)
}

func decodePacket(packet []byte) ([]byte, error) {
	// Decode the BYOND-specific world/Topic() format
	if len(packet) < 4 {
		return nil, fmt.Errorf("packet less than 4 bytes, can't decode, full message: [% x]", packet)
	}
	if len(packet) > 0 && (packet[0] == 0x00 || packet[1] == 0x83) {
		size := binary.BigEndian.Uint16(packet[2:4]) - 1
		if packet[4] == 0x2a { // floating-point number response
			fValue := math.Float32frombits(binary.LittleEndian.Uint32(packet[5:9]))
			return []byte(fmt.Sprintf("%f", fValue)), nil
		} else if packet[4] == 0x06 { // ASCII string response
			// size includes type byte, so subtract 1 from the length
			return packet[5 : 5+size-1], nil
		} else if packet[4] == 0x00 { // world not initialized, empty response or a runtime
			return nil, nil
		}
	}
	return nil, fmt.Errorf("unknown packet type: %#x, full message: [% x]", packet[4], packet)
}
