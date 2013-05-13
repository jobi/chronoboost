package chronoboost

import (
	"fmt"
	"io"
	"net"
)

var timer = 0

func (state *CurrentState) WakeServerUp(conn net.Conn) error {
	if timer++; timer < 10 {
		return nil
	}

	buffer := make([]byte, 1)
	buffer[0] = 0x10
	_, err := conn.Write(buffer)
	if err != nil {
		fmt.Println("Error sending byte to server", err)
		return err
	}

	timer = 0
	return nil
}

func (state *EventState) ReadPacket(reader io.Reader) (*Packet, error) {
	bytes := make([]byte, 2)

	totalRead := 0
	for totalRead < 2 {
		nRead, err := reader.Read(bytes[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += nRead
	}

	header := NewHeader(bytes)

	packet := new(Packet)
	packet.Car = header.GetCar()
	packet.Type = header.GetPacketType()

	var decrypt bool

	if packet.Car != 0 {
		switch CarPacketType(packet.Type) {
		case CAR_POSITION_UPDATE:
			packet.Len = header.GetSpecialPacketLen()
			packet.Data = header.GetSpecialPacketData()
			decrypt = false
		case CAR_POSITION_HISTORY:
			packet.Len = header.GetLongPacketLen()
			packet.Data = header.GetLongPacketData()
			decrypt = true
		default:
			packet.Len = header.GetShortPacketLen()
			packet.Data = header.GetShortPacketData()
			decrypt = true
		}
	} else {
		switch SystemPacketType(packet.Type) {
		case SYS_EVENT_ID:
			fallthrough
		case SYS_KEY_FRAME:
			packet.Len = header.GetShortPacketLen()
			packet.Data = header.GetShortPacketData()
			decrypt = false

		case SYS_TIMESTAMP:
			packet.Len = 2
			packet.Data = 0
			decrypt = true

		case SYS_WEATHER:
			fallthrough
		case SYS_TRACK_STATUS:
			packet.Len = header.GetShortPacketLen()
			packet.Data = header.GetShortPacketData()
			decrypt = true

		case SYS_COMMENTARY:
			fallthrough
		case SYS_NOTICE:
			fallthrough
		case SYS_SPEED:
			packet.Len = header.GetLongPacketLen()
			packet.Data = header.GetLongPacketData()
			decrypt = true

		case SYS_COPYRIGHT:
			packet.Len = header.GetLongPacketLen()
			packet.Data = header.GetLongPacketData()
			decrypt = false

		case SYS_VALID_MARKER:
			fallthrough
		case SYS_REFRESH_RATE:
			packet.Len = 0
			packet.Data = 0
			decrypt = false

		default:
			fmt.Println("Unknown packet type", packet.Type)
			packet.Len = 0
			packet.Data = 0
			decrypt = false
		}
	}

	if packet.Len > 0 {
		packet.Payload = make([]byte, packet.Len)

		totalRead = 0
		for totalRead < packet.Len {
			nRead, err := reader.Read(packet.Payload[totalRead:])

			if err != nil {
				fmt.Println("Got error", err)
				return nil, err
			}

			totalRead += nRead
		}

		if decrypt {
			state.Decrypt(packet)
		}
	} else {
		packet.Payload = nil
	}

	return packet, nil
}
