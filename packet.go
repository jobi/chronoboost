package chronoboost

import (
	"fmt"
)

type PacketType uint8
type CarPacketType PacketType

const (
	CAR_POSITION_UPDATE  CarPacketType = 0
	CAR_POSITION_HISTORY               = 15
	LAST_CAR_PACKET
)

type RaceAtomType uint8

const (
	_ RaceAtomType = iota
	RACE_POSITION
	RACE_NUMBER
	RACE_DRIVER
	RACE_GAP
	RACE_INTERVAL
	RACE_LAP_TIME
	RACE_SECTOR_1
	RACE_PIT_LAP_1
	RACE_SECTOR_2
	RACE_PIT_LAP_2
	RACE_SECTOR_3
	RACE_PIT_LAP_3
	RACE_NUM_PITS
	LAST_RACE_ATOM
)

type PracticeAtomType uint8

const (
	_ PracticeAtomType = iota
	PRACTICE_POSITION
	PRACTICE_NUMBER
	PRACTICE_DRIVER
	PRACTICE_BEST
	PRACTICE_GAP
	PRACTICE_SECTOR_1
	PRACTICE_SECTOR_2
	PRACTICE_SECTOR_3
	PRACTICE_LAP
	LAST_PRACTICE
)

type QualifyingAtomType uint8

const (
	_ QualifyingAtomType = iota
	QUALIFYING_POSITION
	QUALIFYING_NUMBER
	QUALIFYING_DRIVER
	QUALIFYING_PERIOD_1
	QUALIFYING_PERIOD_2
	QUALIFYING_PERIOD_3
	QUALIFYING_SECTOR_1
	QUALIFYING_SECTOR_2
	QUALIFYING_SECTOR_3
	QUALIFYING_LAP
	LAST_QUALIFYING
)

type SystemPacketType PacketType

const (
	_ SystemPacketType = iota
	SYS_EVENT_ID
	SYS_KEY_FRAME
	SYS_VALID_MARKER
	SYS_COMMENTARY
	SYS_REFRESH_RATE
	SYS_NOTICE
	SYS_TIMESTAMP
	_
	SYS_WEATHER
	SYS_SPEED
	SYS_TRACK_STATUS
	SYS_COPYRIGHT
	LAST_SYSTEM_PACKET
)

type WeatherPacketType PacketType

const (
	WEATHER_SESSION_CLOCK WeatherPacketType = iota
	WEATHER_TRACK_TEMP
	WEATHER_AIR_TEMP
	WEATHER_WET_TRACK
	WEATHER_WIND_SPEED
	WEATHER_HUMIDITY
	WEATHER_PRESSURE
	WEATHER_WIND_DIRECTION
)

type SpeedPacketType PacketType

const (
	_ SpeedPacketType = iota
	SPEED_SECTOR1
	SPEED_SECTOR2
	SPEED_SECTOR3
	SPEED_TRAP
	FL_CAR
	FL_DRIVER
	FL_TIME
	FL_LAP
)

type CarNumber uint8

type Packet struct {
	Car     CarNumber
	Type    PacketType
	Data    int
	Len     int
	Payload []byte
}

type PacketHandler interface {
	HandlePacket(packet *Packet, channel chan Object) Object
}

func (packet *Packet) GetDecimalPayload() uint {
	var number uint

	for i := 0; i < packet.Len; i++ {
		if packet.Payload[i] != '.' {
			number *= 10
			number += uint(packet.Payload[i] - '0')
		}
	}

	return number
}

type Header [2]byte

func NewHeader(bytes []byte) *Header {
	header := new(Header)
	header[0] = bytes[0]
	header[1] = bytes[1]
	return header
}

func (header *Header) GetCar() CarNumber {
	return CarNumber(header[0] & 0x1f)
}

func (header *Header) GetPacketType() PacketType {
	return PacketType((header[0] >> 5) | (header[1]&0x01)<<3)
}

func (header *Header) GetSpecialPacketLen() int {
	return 0
}

func (header *Header) GetSpecialPacketData() int {
	return int(header[1] >> 1)
}

func (header *Header) GetShortPacketLen() int {
	if header[1]&0xf0 == 0xf0 {
		return -1
	}

	return int(header[1] >> 4)
}

func (header *Header) GetShortPacketData() int {
	return int((header[1] & 0x0e) >> 1)
}

func (header *Header) GetLongPacketLen() int {
	return int(header[1] >> 1)
}

func (header *Header) GetLongPacketData() int {
	return int(0)
}

func (state *EventState) handleSystemPacket(packet *Packet) Object {
	var number uint = 0
	var result Object = nil

	switch SystemPacketType(packet.Type) {
	case SYS_EVENT_ID:
		for i := 1; i < packet.Len; i++ {
			number *= 10
			number += uint(packet.Payload[i] - '0')
		}

		fmt.Println("Got event", number)

		state.Event = Event{number, EventType(packet.Data)}
		state.Flag = GREEN_FLAG
		state.Reset()

		result = &state.Event

	case SYS_KEY_FRAME:
		for i := packet.Len - 1; i >= 0; i-- {
			number <<= 8
			number |= uint(packet.Payload[i])
		}

		fmt.Println("Got keyframe", number)
		result = KeyFrameNumber(number)

	case SYS_TRACK_STATUS:
		switch packet.Data {
		case 1:
			result = FlagStatus(packet.Payload[0] - '0')
		default:
			fmt.Println("Unhandled flag")
		}

	case SYS_WEATHER:
		switch WeatherPacketType(packet.Data) {
		case WEATHER_TRACK_TEMP:
			state.Weather.TrackTemp = packet.GetDecimalPayload()
		case WEATHER_AIR_TEMP:
			state.Weather.AirTemp = packet.GetDecimalPayload()
		case WEATHER_WIND_SPEED:
			state.Weather.WindSpeed = packet.GetDecimalPayload()
		case WEATHER_WIND_DIRECTION:
			state.Weather.WindDirection = packet.GetDecimalPayload()
		}

		result = state.Weather
	}

	return result
}

func (state *EventState) handleRaceCarAtom(packet *Packet, car *Car) {
	switch RaceAtomType(packet.Type) {
	case RACE_POSITION:
	case RACE_NUMBER:
		if len(packet.Payload) > 0 {
			car.RaceNumber = string(packet.Payload)
		}
	case RACE_DRIVER:
		if len(packet.Payload) > 0 {
			car.Driver = string(packet.Payload)
		}
	case RACE_GAP:
		if len(packet.Payload) > 0 {
			car.Gap = string(packet.Payload)
		}
	case RACE_INTERVAL:
		if len(packet.Payload) > 0 {
			car.Interval = string(packet.Payload)
		}
	case RACE_LAP_TIME:
		if len(packet.Payload) > 0 {
			car.LapTime = string(packet.Payload)
		}
		car.LapTimeStatus = TimeStatus(packet.Data)
	case RACE_SECTOR_1:
		if len(packet.Payload) > 0 {
			car.Sector1 = string(packet.Payload)
		}
		car.Sector1Status = TimeStatus(packet.Data)
	case RACE_SECTOR_2:
		if len(packet.Payload) > 0 {
			car.Sector2 = string(packet.Payload)
		}
		car.Sector2Status = TimeStatus(packet.Data)
	case RACE_SECTOR_3:
		if len(packet.Payload) > 0 {
			car.Sector3 = string(packet.Payload)
		}
		car.Sector3Status = TimeStatus(packet.Data)
	}
}

func (state *EventState) handlePracticeCarAtom(packet *Packet, car *Car) {
	switch PracticeAtomType(packet.Type) {
	case PRACTICE_POSITION:
	case PRACTICE_NUMBER:
		car.RaceNumber = string(packet.Payload)
	case PRACTICE_DRIVER:
		car.Driver = string(packet.Payload)
	case PRACTICE_GAP:
		car.Gap = string(packet.Payload)
	case PRACTICE_BEST:
		car.LapTime = string(packet.Payload)
	case PRACTICE_SECTOR_1:
		car.Sector1 = string(packet.Payload)
	case PRACTICE_SECTOR_2:
		car.Sector2 = string(packet.Payload)
	case PRACTICE_SECTOR_3:
		car.Sector3 = string(packet.Payload)
	}
}

func (state *EventState) handleQualifyingCarAtom(packet *Packet, car *Car) {
	switch QualifyingAtomType(packet.Type) {
	case QUALIFYING_POSITION:
	case QUALIFYING_NUMBER:
		car.RaceNumber = string(packet.Payload)
	case QUALIFYING_DRIVER:
		car.Driver = string(packet.Payload)
	case QUALIFYING_PERIOD_1:
		fallthrough
	case QUALIFYING_PERIOD_2:
		fallthrough
	case QUALIFYING_PERIOD_3:
		car.LapTime = string(packet.Payload)
		car.LapTimeStatus = TimeStatus(packet.Data)
	case QUALIFYING_SECTOR_1:
		car.Sector1 = string(packet.Payload)
		car.Sector1Status = TimeStatus(packet.Data)
	case QUALIFYING_SECTOR_2:
		car.Sector2 = string(packet.Payload)
		car.Sector2Status = TimeStatus(packet.Data)
	case QUALIFYING_SECTOR_3:
		car.Sector3 = string(packet.Payload)
		car.Sector3Status = TimeStatus(packet.Data)
	}
}

func (state *EventState) handleCarPacket(packet *Packet) *Car {
	car := state.Cars[packet.Car]
	if car == nil {
		car = new(Car)
		car.Number = packet.Car
		state.Cars[packet.Car] = car
	}

	switch CarPacketType(packet.Type) {
	case CAR_POSITION_UPDATE:
		for number, carInfo := range state.Cars {
			if number == packet.Car {
				carInfo.Position = packet.Data
			} else if carInfo.Position == packet.Data {
				carInfo.Position = 0
			}
		}
	case CAR_POSITION_HISTORY:
	default:
		switch state.Event.Type {
		case RACE_EVENT:
			state.handleRaceCarAtom(packet, car)
		case PRACTIVE_EVENT:
			state.handlePracticeCarAtom(packet, car)
		case QUALIFYING_EVENT:
			state.handleQualifyingCarAtom(packet, car)
		}
	}

	return car
}

func (state *EventState) HandlePacket(packet *Packet, channel chan Object) Object {
	var result Object

	if packet.Car == 0 {
		result = state.handleSystemPacket(packet)
	} else {
		result = state.handleCarPacket(packet)
	}

	return result
}
