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

type analyticsEvent struct {
	eventCollection string
	data            interface{}
}

type Sender struct {
	projectId string
	writeKey  string
	url       string
	events    map[string][]interface{}
	count     int
	channel   chan analyticsEvent
}

/*
Create a new Sender.

This creates a background goroutine to aggregate and send your events.
*/
func NewSender(projectId, writeKey string) *Sender {
	ae := &Sender{
		projectId: projectId,
		writeKey:  writeKey,
		channel:   make(chan analyticsEvent, channel_size),
	}
	ae.url = url + ae.projectId + "/events?api_key=" + ae.writeKey
	ae.reset()
	go ae.run()
	return ae
}

/*
Queue events to be sent to Keen.io

info can be anything that is JSON serializable.  Events are immediately queued to a background goroutine for sending.  The
background routine will send everything that's queued to it in a batch, then wait for new data.

The upshot is that if you send events slowly they will be sent immediately and individually, but if you send events quickly they will be batched
*/
func (ae *Sender) Queue(eventCollection string, info interface{}) {
	ae.channel <- analyticsEvent{eventCollection, info}
}

func (ae *Sender) add(event analyticsEvent) {
	ae.events[event.eventCollection] = append(ae.events[event.eventCollection], event.data)
	ae.count++

	if ae.count > send_threshold {
		ae.send()
	}
}

func (ae *Sender) reset() {
	ae.events = make(map[string][]interface{}, 10)
	ae.count = 0
}

func (ae *Sender) send() {
	// Send the events currently in ae.events
	if ae.count == 0 {
		return
	}
	defer ae.reset()

	// Convert data to JSON
	data, err := json.Marshal(ae.events)
	if err != nil {
		log.Printf("Couldn't marshal json for analytics. %v\n", err)
		return
	}

	start := time.Now()
	rsp, err := http.Post(ae.url, "application/json", strings.NewReader(string(data)))
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

func (ae *Sender) run() {
	var event analyticsEvent

	// Block for the first event, once we have one event we try to drain everthing left
	for event = range ae.channel {
		ae.add(event)

		// Select with a default case is essentially a non-blocking read from the channel
	Loop:
		for {
			select {
			case event = <-ae.channel:
				// Add the event to those we are batching
				ae.add(event)

			default:
				// Nothing to batch at present.  Send our events if we have any, then go back to block until something
				// shows up
				ae.send()
				break Loop
			}
		}
	}
}
