/*
Send events to Keen.io asynchronously in batches.
*/
package keengo

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	// The Keen.IO API URL
	url string = "https://api.keen.io/3.0/projects/"
	// The size of the queue to the background goroutine
	channel_size int = 100
	// The background routine will send batches of events up to this size
	send_threshold int = 90
)

// Type for queuing events to the background
type analyticsEvent struct {
	eventCollection string
	data            interface{}
}

type Sender struct {
	projectId string
	writeKey  string
	url       string                   // The url to post events too, including project details
	events    map[string][]interface{} // For batching events as we pull them off the channel
	count     int                      // Number of events batched and ready to send
	channel   chan analyticsEvent      // For queuing events to the background
	done      chan bool                // For clean exiting
}

/*
Create a new Sender.

This creates a background goroutine to aggregate and send your events.
*/
func NewSender(projectId, writeKey string) *Sender {
	sender := &Sender{
		projectId: projectId,
		writeKey:  writeKey,
		channel:   make(chan analyticsEvent, channel_size),
		done:      make(chan bool),
	}
	sender.url = url + sender.projectId + "/events?api_key=" + sender.writeKey
	sender.reset()
	go sender.run()
	return sender
}

/*
Queue events to be sent to Keen.io

info can be anything that is JSON serializable.  Events are immediately queued to a background goroutine for sending.  The
background routine will send everything that's queued to it in a batch, then wait for new data.

The upshot is that if you send events slowly they will be sent immediately and individually, but if you send events quickly they will be batched
*/
func (sender *Sender) Queue(eventCollection string, info interface{}) {
	sender.channel <- analyticsEvent{eventCollection, info}
}

/*
Close the sender and wait for queued events to be sent
*/
func (sender *Sender) Close() {
	// Closing the channel signals the background thread to exit
	close(sender.channel)
	// Wait for the background thread to signal it has flushed all events and exited
	<-sender.done
}

// Add an event to the map that's used to batch events
func (sender *Sender) add(event analyticsEvent) bool {
	if event.eventCollection == "" {
		// nil event, don't add
		return false
	}
	sender.events[event.eventCollection] = append(sender.events[event.eventCollection], event.data)
	sender.count++

	if sender.count > send_threshold {
		sender.send()
	}
	return true
}

// Reset the event map that's used to batch events
func (sender *Sender) reset() {
	sender.events = make(map[string][]interface{}, 10)
	sender.count = 0
}

// Send the events currently in sender.events
func (sender *Sender) send() {
	if sender.count == 0 {
		return
	}
	// Whether we can send the events or not, we dump them before exiting this function
	defer sender.reset()

	// Convert data to JSON
	data, err := json.Marshal(sender.events)
	if err != nil {
		log.Printf("Couldn't marshal json for analytics. %v\n", err)
		return
	}

	start := time.Now()
	rsp, err := http.Post(sender.url, "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Printf("Failed to post analytics events.  %v\n", err)
		return
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		log.Printf("Failure return for analytics post.  %d, %s\n", rsp.StatusCode, rsp.Status)
	} else {
		// TODO: remove once analytics has bedded in
		log.Printf("analytics sent in %v\n", time.Since(start))
	}
}

func (sender *Sender) run() {
	var event analyticsEvent

	// Block for the first event, once we have one event we try to drain everthing left
	for event = range sender.channel {
		sender.add(event)

		// Select with a default case is essentially a non-blocking read from the channel
	Loop:
		for {
			select {
			case event = <-sender.channel:
				// Add the event to those we are batching
				if !sender.add(event) {
					break Loop
				}

			default:
				// Nothing to batch at present.  Send our events if we have any, then go back to block until something
				// shows up
				break Loop
			}
		}
		// Send what we have batched
		sender.send()
	}

	// Indicate that this thread is over
	sender.done <- true
	log.Printf("Analytics exited\n")
}
