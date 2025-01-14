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
	"io"

	"go.chromium.org/luci/server"
	"go.chromium.org/luci/server/router"

	"go.chromium.org/luci/resultdb/internal"
	pb "go.chromium.org/luci/resultdb/proto/rpc/v1"
)

func main() {
	internal.Main(func(srv *server.Server) error {
		srv.Routes.GET("/", router.MiddlewareChain{}, func(c *router.Context) {
			io.WriteString(c.Writer, "OK")
		})
		pb.RegisterRecorderServer(srv.PRPC, NewRecorderServer())
		return nil
	})
}
