package main

import (
	"sync"
	"fmt"
	"os"
	"os/signal"
	"net/url"
	"log"
	"github.com/gorilla/websocket"
	"time"
)

/**
Websocket client using a basic supervision pattern, so that in the event that the websocket connection breaks, the
system schedules another connection.
 */

func websocketRoutine(deadLetters chan string, connected chan bool) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
		connected <- false
		deadLetters <- "websocket disconnected"
	}()
	const addr = "localhost:8085"
	const path = "/connect"
	const schema = "ws"

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: schema, Host: addr, Path: path}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Panicf("dial: %s", err)
	}
	connected <- true

	defer func() {
		c.Close()
	}()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)

	for {
		select {
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
		// To cleanly close a connection, a client should send a close
		// frame and wait for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
				select {
				case <-done:
				case <-time.After(time.Second):
				}
			c.Close()
			return
		}
	}
}

func supervisor() {
	deadletters := make(chan string,10)
	connected := make(chan bool, 10)
	defer func() {
		wg.Done()
	}()

	go websocketRoutine(deadletters, connected)

	//wait for events to occur.

	for {
		select {
			case deadLetter := <-deadletters:
				log.Printf("\nSupervisor: got dead letter from websocket, rescheduling %s\n", deadLetter)
				time.Sleep(2 * time.Second)
				go websocketRoutine(deadletters, connected)

			case c := <- connected:
				log.Printf("Supervisor: got connected: %d from websocket", c)
		}
	}
}

var wg sync.WaitGroup

func main() {
	wg.Add(1)
	go supervisor()
	fmt.Println("Waiting for supervisor to die")
	wg.Wait()
	fmt.Println("Bye")
}
