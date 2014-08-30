#keengo

Send events to [Keen.io](https://keen.io) asynchronously in batches.

See https://github.com/philpearl/keengo_goji for Goji middleware that sends HTTP request/response info to keen.io.

See the documentation at https://godoc.org/github.com/philpearl/keengo

## Example

```golang

package main

import (
        "fmt"
        "net/http"

        "github.com/philpearl/keengo"
)


func main() {
    // You just need one sender object - feed it your keen project ID and write key
    sender := keengo.NewSender(projectId, writeKey)

    // The data you send can be anything JSON serialisable
    info := map[string]interface{}{
        "data": "hello world",
    }

    // Call Queue to send your event
    sender.Queue("my first collection", info)
}

```
