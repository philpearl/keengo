/*
Middleware for Goji that reports HTTP requests to Keen.io.

Add it to your Goji mux m as follows.

	m.Use(BuildMiddleWare(myKeenIoProjectId, myKeenIoWriteKey, "apievents", nil))

To add your own data to the events reported add a callback.

	callback := func(c *web.C, data map[string]interface{}, r *http.Request) {
		data["my_parameter"] = c.Env["important_info"].(string)
	}

	m.Use(BuildMiddleWare(myKeenIoProjectId, myKeenIoWriteKey, "apievents", callback))

*/
package goji_middleware

import (
	"net/http"
	"time"

	"github.com/philpearl/keengo"
	"github.com/zenazn/goji/web"
)

/*
Build a middleware function that reports HTTP requests to Keen.io.

The event the middleware sends is as follows.

	{
		"url":         r.RequestURI,
		"path":        r.URL.Path,
		"method":      r.Method,
		"status_code": w.Status,  // Http status code from response
		"duration_ns": time.Since(start).Nanoseconds(),
		"user_agent":  r.UserAgent(),
		"header":      r.Header,
	}

The callback function allows you to add your own data to the event recorded.  Use it as follows to add events to the "apievents" collection.

	callback := func(c *web.C, data map[string]interface{}, r *http.Request) {
		data["my_parameter"] = c.Env["important_info"].(string)
	}

	m.Use(BuildMiddleWare(myKeenIoProjectId, myKeenIoWriteKey, "apievents", callback))

*/
func BuildMiddleWare(projectId, writeKey, collectionName string,
	callback func(c *web.C, data map[string]interface{}, r *http.Request),
) func(c *web.C, h http.Handler) http.Handler {
	sender := keengo.NewSender(projectId, writeKey)

	// Return the middleware that references the analytics queue we just made
	return func(c *web.C, h http.Handler) http.Handler {
		handler := func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := &StatusTrackingResponseWriter{w, http.StatusOK}

			h.ServeHTTP(ww, r)

			info := map[string]interface{}{
				"url":         r.RequestURI,
				"path":        r.URL.Path,
				"method":      r.Method,
				"status_code": ww.Status,
				"duration_ns": time.Since(start).Nanoseconds(),
				"user_agent":  r.UserAgent(),
				"header":      r.Header,
			}
			// Get more data for the analytics event
			if callback != nil {
				callback(c, info, r)
			}

			sender.Queue(collectionName, info)
		}
		return http.HandlerFunc(handler)
	}
}
