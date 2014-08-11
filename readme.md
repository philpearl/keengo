#keengo

Send events to [Keen.io](http://keen.io) asynchronously in batches.

Includes middleware for [Goji](https://github.com/zenazn/goji) to track HTTP requests and responses.

See the documentation at https://godoc.org/github.com/philpearl/keengo

## Goji Middleware Example
You can use keengo to send individual events, but I'm mostly using the included Goji middleware to track calls to my API.  Here's how.

```golang

package main

import (
        "fmt"
        "net/http"

        keengo "github.com/philpearl/keengo/goji_middleware"

        "github.com/zenazn/goji"
        "github.com/zenazn/goji/web"
)

func hello(c web.C, w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %s!", c.URLParams["name"])
}

func main() {
        goji.Use(keengo.BuildMiddleWare("mykeenprojectID", "mykeenwritekey","requests", nil)
        goji.Get("/hello/:name", hello)
        goji.Serve()
}

```
