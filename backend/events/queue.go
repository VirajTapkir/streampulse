package events

import (
	"math/rand"
	"time"
)

// Event represents a single Twitch-style monetization event
type Event struct {
	Type      string  // "sub", "bits", or "donation"
	Username  string  // who triggered the event
	Amount    float64 // dollar value
	Timestamp time.Time
}

// eventTypes is the list of possible events we randomly pick from
var eventTypes = []string{"sub", "bits", "donation"}

// usernames is a pool of fake usernames to make events feel realistic
var usernames = []string{
	"xX_gamer_Xx", "streamfan99", "nightowl42",
	"coolviewer", "hyperchatter", "lurker2000",
}

// StartEventQueue launches a goroutine that sends a fake event
// every 1-3 seconds into the returned channel
func StartEventQueue() chan Event {
	// make a buffered channel that can hold up to 100 events
	// without blocking if nothing is reading from it yet
	eventChan := make(chan Event, 100)

	go func() {
		for {
			// pick a random event type and username
			evt := Event{
				Type:      eventTypes[rand.Intn(len(eventTypes))],
				Username:  usernames[rand.Intn(len(usernames))],
				Amount:    randomAmount(),
				Timestamp: time.Now(),
			}

			// send the event into the channel
			eventChan <- evt

			// wait between 1 and 3 seconds before the next event
			delay := time.Duration(1+rand.Intn(3)) * time.Second
			time.Sleep(delay)
		}
	}()

	return eventChan
}

// randomAmount returns a realistic dollar amount based on event type
// we call this separately so the logic is easy to read
func randomAmount() float64 {
	amounts := []float64{4.99, 9.99, 24.99, 100.00, 5.00, 10.00}
	return amounts[rand.Intn(len(amounts))]
}
