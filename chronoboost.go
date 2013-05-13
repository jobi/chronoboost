package chronoboost

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"time"
)

type EventType uint8

const (
	_ EventType = iota
	RACE_EVENT
	PRACTIVE_EVENT
	QUALIFYING_EVENT
)

type FlagStatus uint8

const (
	_ FlagStatus = iota
	GREEN_FLAG
	YELLOW_FLAG
	SAFETY_CAR_STANDBY
	SAFETY_CAR_DEPLOYED
	RED_FLAG
	LAST_FLAG
)

func (flagStatus FlagStatus) String() string {
	switch flagStatus {
	case GREEN_FLAG:
		return "Green flag"
	case YELLOW_FLAG:
		return "Yellow flag"
	case RED_FLAG:
		return "Red flag"
	case SAFETY_CAR_DEPLOYED:
		return "Safety car deployed"
	case SAFETY_CAR_STANDBY:
		return "Safety car standby"
	}

	return "Unknown flag"
}

type Weather struct {
	TrackTemp     uint
	AirTemp       uint
	Humidity      uint
	WindSpeed     uint
	WindDirection uint
	Pressure      uint
}

func (weather Weather) String() string {
	return fmt.Sprintf("Track temperature: %d - Air temperature: %d", weather.TrackTemp, weather.AirTemp)
}

type TimeStatus uint8

const (
	Normal TimeStatus = iota
	Latest
	Pit
	PersonalBest
	Record
	_
	Old
)

type Object interface {
	String() string
}

type Car struct {
	Number        CarNumber
	RaceNumber    string
	Position      int
	Driver        string
	Gap           string
	Interval      string
	LapTime       string
	LapTimeStatus TimeStatus
	Sector1       string
	Sector1Status TimeStatus
	Sector2       string
	Sector2Status TimeStatus
	Sector3       string
	Sector3Status TimeStatus
}

func (car Car) String() string {
	return fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", car.Position, car.RaceNumber, car.Driver, car.Gap, car.Interval, car.LapTime, car.Sector1, car.Sector2, car.Sector3)
}

type Event struct {
	Number uint
	Type   EventType
}

func (event Event) String() string {
	return fmt.Sprintf("Event %u", event.Number)
}

type KeyFrameNumber uint

func (frame KeyFrameNumber) String() string {
	return fmt.Sprintf("Keyframe %u", frame)
}

type Decrypter interface {
	Decrypt(packet *Packet)
	Reset()
	SetKey(key uint)
}

type CryptoState struct {
	Key               uint
	Salt              uint
	DecryptionFailure bool
}

func (state *CryptoState) SetKey(key uint) {
	state.Key = key
}

func (state *CryptoState) Decrypt(packet *Packet) {
	if state.Key == 0 {
		return
	}

	for i := 0; i < packet.Len; i++ {
		var mask uint
		if state.Salt&0x01 != 0 {
			mask = state.Key
		} else {
			mask = 0
		}

		state.Salt = ((state.Salt >> 1) ^ mask)
		packet.Payload[i] = packet.Payload[i] ^ byte(state.Salt&0xff)
	}
}

const CRYPTO_SEED = 0x55555555

func (state *CryptoState) Reset() {
	state.Salt = CRYPTO_SEED
}

func NewCryptoState() *CryptoState {
	state := new(CryptoState)
	state.Reset()

	return state
}

type EventState struct {
	*CryptoState
	*RacingState
}

func NewEventState() *EventState {
	state := new(EventState)

	state.CryptoState = NewCryptoState()
	state.RacingState = NewRacingState()

	return state
}

type RacingState struct {
	Event   Event
	Cars    map[CarNumber]*Car
	Weather *Weather
	Flag    FlagStatus
}

func NewRacingState() *RacingState {
	state := new(RacingState)

	state.Event = Event{0, RACE_EVENT}
	state.Flag = GREEN_FLAG
	state.Cars = make(map[CarNumber]*Car)
	state.Weather = new(Weather)

	return state
}

type CurrentState struct {
	Host               string
	Port               int
	AuthHost           string
	Email              string
	Password           string
	Token              string
	LastKeyFrameNumber KeyFrameNumber
	LastKeyFrame       *[]byte
	ParsingKeyFrame    bool
	*EventState
}

func NewCurrentState() *CurrentState {
	state := new(CurrentState)

	state.EventState = NewEventState()

	return state
}

func (state *CurrentState) HandlePacket(packet *Packet, channel chan Object) Object {
	obj := state.EventState.HandlePacket(packet, channel)

	switch obj.(type) {
	case *Event:
		state.ObtainDecryptionKey()
		state.Reset()

	case KeyFrameNumber:
		state.Reset()

		if !state.ParsingKeyFrame {
			keyFrameNumber := obj.(KeyFrameNumber)
			if state.LastKeyFrameNumber != keyFrameNumber {
				frame, err := state.ObtainKeyFrame(keyFrameNumber)
				if err == nil {
					state.ParsingKeyFrame = true
					state.playKeyFrame(state, &frame, channel)
					state.ParsingKeyFrame = false
					state.Reset()

					state.LastKeyFrameNumber = keyFrameNumber
					state.LastKeyFrame = &frame
				}
			}
		}
	}

	return obj
}

func (state *CurrentState) mainLoop(channel chan Object) {
	for {
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", state.Host, state.Port))
		if err != nil {
			panic(fmt.Sprintf("Failed to open connection: %s", err))
		}
		defer conn.Close()

		go func() {
			for {
				state.WakeServerUp(conn)
				time.Sleep(1 * time.Second)
			}

		}()

		state.Reset()
		reader := bufio.NewReader(conn)

		for {
			packet, err := state.ReadPacket(reader)
			if err != nil {
				break
			}

			if packet == nil {
				break
			}

			obj := state.HandlePacket(packet, channel)

			if obj != nil {
				channel <- obj
			}
		}

		if err != nil {
			panic(fmt.Sprintf("Error reading: %s", err))
		}
	}
}

func (state *CurrentState) ReplayLastFrame(channel chan Object) {
	if state.LastKeyFrame == nil {
		fmt.Println("No last frame")
		return
	}

	cryptoState := NewCryptoState()
	cryptoState.Key = state.Key

	eventState := NewEventState()
	eventState.CryptoState = cryptoState
	cryptoState.Reset()

	eventState.playKeyFrame(eventState, state.LastKeyFrame, channel)
}

func (state *EventState) playKeyFrame(packetHandler PacketHandler, keyframe *[]byte, channel chan Object) {
	reader := bytes.NewBuffer(*keyframe)
	for {
		packet, err := state.ReadPacket(reader)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("Error reading packet", err)
			break
		}

		if packet == nil {
			break
		}

		obj := packetHandler.HandlePacket(packet, channel)

		if obj != nil {
			channel <- obj
		}
	}
}

func (state *CurrentState) Run(channel chan Object) {
	go state.mainLoop(channel)
}
