package keengo_test

import (
	"github.com/philpearl/keengo"
	"log"
	"testing"
)

// TODO: pull these from somewhere for testing
const projectId = ""
const writeKey = ""

func TestSend(t *testing.T) {
	log.Printf("start\n")
	sender := keengo.NewSender(projectId, writeKey)
	log.Printf("Have a sender\n")

	sender.Queue("testing", map[string]interface{}{
		"hat":    1,
		"cheese": "wensleydale",
	})

	log.Printf("Queued\n")

	sender.Close()
}
