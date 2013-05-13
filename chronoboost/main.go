package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"

	"code.google.com/p/go.net/websocket"
	"github.com/jobi/chronoboost"
)

const defaultHost = "live-timing.formula1.com"
const defaultPort = 4321

var listeners []chan chronoboost.Object = make([]chan chronoboost.Object, 0)

func dispatchObjects(input chan chronoboost.Object) {
	for {
		obj := <-input

		for _, output := range listeners {
			output <- obj
		}
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Got request")

	t, err := template.ParseFiles("static/html/index.html")
	if err != nil {

	}

	t.Execute(w, nil)
}

func main() {
	state := chronoboost.NewCurrentState()
	state.AuthHost = defaultHost
	state.Host = defaultHost
	state.Port = defaultPort

	flag.StringVar(&state.Email, "email", "", "Email address (formula1.com account)")
	flag.StringVar(&state.Password, "password", "", "Password (formula1.com account)")
	flag.Parse()

	err := state.ObtainAuthCookie()

	if err != nil {
		fmt.Println("Failed to acquire cookie", err)
		return
	}

	websocketHandler := func(ws *websocket.Conn) {
		var msg struct {
			Type  string
			Value chronoboost.Object
		}

		listener := make(chan chronoboost.Object)

		// replay the last key frame then subscribe
		go func() {
			state.ReplayLastFrame(listener)
			listeners = append(listeners, listener)
		}()

		for {
			msg.Value = <-listener

			switch msg.Value.(type) {
			case *chronoboost.Car:
				msg.Type = "car"
			case *chronoboost.FlagStatus:
				msg.Type = "flagStatus"
			case *chronoboost.Weather:
				msg.Type = "weather"
			default:
				continue
			}

			websocket.JSON.Send(ws, msg)
		}
	}

	packetChannel := make(chan chronoboost.Object)
	go dispatchObjects(packetChannel)
	state.Run(packetChannel)

	http.Handle("/ws", websocket.Handler(websocketHandler))
	http.Handle("/static/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/", rootHandler)
	http.ListenAndServe(":8080", nil)
}
