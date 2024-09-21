package ss13

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"net"
)

type ServerStatus struct {
	Players   []string `json:"playerlist"`
	RoundTime string   `json:"roundtime"`
	Map       string   `json:"map"`
	Evac      int      `json:"evac"`
}

func GetServerStatus(serverAddress string, ctx context.Context) (ServerStatus, error) {
	var result ServerStatus
	resp, err := sendRequest(serverAddress, []byte("discordstatus"), ctx)
	if err != nil {
		return result, fmt.Errorf("failed to send request: %w", err)
	}
	err = json.Unmarshal(resp, &result)
	if err != nil {
		return result, fmt.Errorf("failed to unmarshal: %w", err)
	}
	return result, nil
}

func sendRequest(serverAddress string, request []byte, ctx context.Context) ([]byte, error) {
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
	conn, err := dialer.DialContext(ctx, "tcp", serverAddress)
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
		} else if packet[4] == 0x00 { // world hasn't loaded or error, let's be optimistic
			return []byte(`{"roundtime": "World initializing...", "map": "Loading..."}`), nil
		}
	}
	return nil, fmt.Errorf("unknown packet type: %#x, full message: [% x]", packet[4], packet)
}
