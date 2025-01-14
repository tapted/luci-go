// Copyright 2019 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"net/http"
	"time"

	"go.opencensus.io/trace"

	"github.com/gomodule/redigo/redis"
	"go.chromium.org/luci/common/logging"

	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/redisconn"
	"go.chromium.org/luci/server/router"
)

func main() {
	server.Main(nil, func(srv *server.Server) error {
		// Logging example.
		srv.Routes.GET("/", router.MiddlewareChain{}, func(c *router.Context) {
			logging.Debugf(c.Context, "Hello debug world")

			ctx, span := trace.StartSpan(c.Context, "Testing")
			logging.Infof(ctx, "Hello info world")
			time.Sleep(100 * time.Millisecond)
			span.End()

			logging.Warningf(c.Context, "Hello warning world")
			c.Writer.Write([]byte("Hello, world"))
		})

		// Redis example.
		//
		// To run Redis for tests locally (in particular on OSX):
		//   docker run --name redis -p 6379:6379 --restart always --detach redis
		//
		// Then launch the example with "... -redis-addr :6379".
		//
		// Note that it makes Redis port available on 0.0.0.0. This is a necessity
		// when using Docker-for-Mac. Don't put any sensitive stuff there (or make
		// sure your firewall is configured to block external connections).
		srv.Routes.GET("/redis", router.MiddlewareChain{}, func(c *router.Context) {
			conn, err := redisconn.Get(c.Context)
			if err != nil {
				http.Error(c.Writer, err.Error(), 500)
				return
			}
			defer conn.Close()
			n, err := redis.Int(conn.Do("INCR", "testKey"))
			if err != nil {
				http.Error(c.Writer, err.Error(), 500)
				return
			}
			fmt.Fprintf(c.Writer, "%d", n)
		})

		return nil
	})
}
