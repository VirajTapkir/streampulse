package events

import (
	"fmt"
	"math/rand"
	"time"
)

type Event struct {
	StreamerID    int
	Type          string
	Username      string
	Amount        float64
	Timestamp     time.Time
	TwitchPayload map[string]interface{}
}


var subTiers = []string{"1000", "2000", "3000"}

var usernames = []string{
	"xX_gamer_Xx", "streamfan99", "nightowl42",
	"coolviewer", "hyperchatter", "lurker2000",
}

var displayNames = map[string]string{
	"xX_gamer_Xx":  "xX_gamer_Xx",
	"streamfan99":  "StreamFan99",
	"nightowl42":   "NightOwl42",
	"coolviewer":   "CoolViewer",
	"hyperchatter": "HyperChatter",
	"lurker2000":   "Lurker2000",
}

var streamerIDs = []int{1, 3, 4}

func StartEventQueue() chan Event {
	eventChan := make(chan Event, 100)

	go func() {
		for {
			streamerID := streamerIDs[rand.Intn(len(streamerIDs))]

			n := rand.Intn(3)
			var evt Event

			switch n {
			case 0:
				evt = generateSubEvent()
			case 1:
				evt = generateCheerEvent()
			case 2:
				evt = generateDonationEvent()
			}

			evt.StreamerID = streamerID
			evt.TwitchPayload["_meta"].(map[string]interface{})["streamer_id"] = streamerID

			eventChan <- evt

			delay := time.Duration(1+rand.Intn(3)) * time.Second
			time.Sleep(delay)
		}
	}()

	return eventChan
}


func generateSubEvent() Event {
	username := usernames[rand.Intn(len(usernames))]
	tier := subTiers[rand.Intn(len(subTiers))]
	isGift := rand.Intn(5) == 0 // 20% chance of gifted sub

	amount := map[string]float64{
		"1000": 4.99,
		"2000": 9.99,
		"3000": 24.99,
	}[tier]

	payload := map[string]interface{}{
		"subscription": map[string]interface{}{
			"id":      fmt.Sprintf("sub-%d", rand.Int()),
			"type":    "channel.subscribe",
			"version": "1",
			"condition": map[string]interface{}{
				"broadcaster_user_id": "12345678",
			},
		},
		"event": map[string]interface{}{
			"user_id":              fmt.Sprintf("%d", rand.Intn(9000000)+1000000),
			"user_login":           username,
			"user_name":            displayNames[username],
			"broadcaster_user_id":  "12345678",
			"broadcaster_user_login": "teststreamer",
			"broadcaster_user_name": "TestStreamer",
			"tier":                 tier,
			"is_gift":              isGift,
		},
		"_meta": map[string]interface{}{
			"type":      "sub",
			"amount":    amount,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return Event{
		Type:          "sub",
		Username:      username,
		Amount:        amount,
		Timestamp:     time.Now(),
		TwitchPayload: payload,
	}
}

func generateCheerEvent() Event {
	username := usernames[rand.Intn(len(usernames))]

	bitsOptions := []int{100, 500, 1000, 5000, 10000}
	bits := bitsOptions[rand.Intn(len(bitsOptions))]

	// Twitch pays out $0.01 per bit
	amount := float64(bits) * 0.01

	messages := []string{
		fmt.Sprintf("PogChamp %d bits!", bits),
		fmt.Sprintf("Keep it up! %d", bits),
		"Great stream!",
		fmt.Sprintf("Cheering %d bits!", bits),
	}
	message := messages[rand.Intn(len(messages))]

	payload := map[string]interface{}{
		"subscription": map[string]interface{}{
			"id":      fmt.Sprintf("cheer-%d", rand.Int()),
			"type":    "channel.cheer",
			"version": "1",
			"condition": map[string]interface{}{
				"broadcaster_user_id": "12345678",
			},
		},
		"event": map[string]interface{}{
			"is_anonymous":           false,
			"user_id":                fmt.Sprintf("%d", rand.Intn(9000000)+1000000),
			"user_login":             username,
			"user_name":              displayNames[username],
			"broadcaster_user_id":    "12345678",
			"broadcaster_user_login": "teststreamer",
			"broadcaster_user_name":  "TestStreamer",
			"message":                message,
			"bits":                   bits,
		},
		"_meta": map[string]interface{}{
			"type":      "bits",
			"amount":    amount,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return Event{
		Type:          "bits",
		Username:      username,
		Amount:        amount,
		Timestamp:     time.Now(),
		TwitchPayload: payload,
	}
}

func generateDonationEvent() Event {
	username := usernames[rand.Intn(len(usernames))]

	amountOptions := []int{500, 1000, 2500, 5000, 10000}
	amountCents := amountOptions[rand.Intn(len(amountOptions))]
	amount := float64(amountCents) / 100.0

	payload := map[string]interface{}{
		"subscription": map[string]interface{}{
			"id":      fmt.Sprintf("donation-%d", rand.Int()),
			"type":    "channel.charity_campaign.donate",
			"version": "1",
			"condition": map[string]interface{}{
				"broadcaster_user_id": "12345678",
			},
		},
		"event": map[string]interface{}{
			"campaign_id":            fmt.Sprintf("campaign-%d", rand.Intn(1000)),
			"user_id":                fmt.Sprintf("%d", rand.Intn(9000000)+1000000),
			"user_login":             username,
			"user_name":              displayNames[username],
			"broadcaster_user_id":    "12345678",
			"broadcaster_user_login": "teststreamer",
			"broadcaster_user_name":  "TestStreamer",
			"charity_name":           "StreamPulse Charity",
			"charity_logo":           "https://streampulse.example.com/logo.png",
			"amount": map[string]interface{}{
				"value":          amountCents,
				"decimal_places": 2,
				"currency":       "USD",
			},
		},
		"_meta": map[string]interface{}{
			"type":      "donation",
			"amount":    amount,
			"timestamp": time.Now().Format(time.RFC3339),
		},
	}

	return Event{
		Type:          "donation",
		Username:      username,
		Amount:        amount,
		Timestamp:     time.Now(),
		TwitchPayload: payload,
	}
}