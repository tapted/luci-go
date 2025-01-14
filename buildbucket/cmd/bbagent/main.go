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

// Command bbagent is Buildbucket's agent running in swarming.
//
// This executable creates a luciexe 'host' environment, and runs the
// Buildbucket build's exe within this environment. Please see
// https://go.chromium.org/luci/luciexe for details about the 'luciexe'
// protocol.
//
// This command is an implementation detail of Buildbucket.
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/golang/protobuf/jsonpb"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	"go.chromium.org/luci/common/logging/gologger"
	"go.chromium.org/luci/lucictx"
	"go.chromium.org/luci/luciexe/host"
	"go.chromium.org/luci/luciexe/invoke"

	"go.chromium.org/luci/buildbucket/cmd/bbagent/bbinput"
	bbpb "go.chromium.org/luci/buildbucket/proto"
)

func main() {
	os.Exit(mainImpl())
}

func mainImpl() int {
	ctx := logging.SetLevel(gologger.StdConfig.Use(context.Background()), logging.Info)

	check := func(err error) {
		if err != nil {
			logging.Errorf(ctx, err.Error())
			os.Exit(1)
		}
	}

	if len(os.Args) != 2 {
		check(errors.Reason("expected 1 argument after arg0, got %d", len(os.Args)-1).Err())
	}

	input, err := bbinput.Parse(os.Args[1])
	check(errors.Annotate(err, "could not unmarshal BBAgentArgs").Err())

	sctx, err := lucictx.SwitchLocalAccount(ctx, "system")
	check(errors.Annotate(err, "could not switch to 'system' account in LUCI_CONTEXT").Err())

	bbClient, err := newBuildsClient(sctx, input.Build.Infra.Buildbucket)
	check(errors.Annotate(err, "could not connect to Buildbucket").Err())
	defer bbClient.CloseAndDrain(ctx)

	// from this point forward we want to try to report errors to buildbucket,
	// too.
	check = func(err error) {
		if err != nil {
			logging.Errorf(ctx, err.Error())
			bbClient.C <- &bbpb.Build{
				Status:          bbpb.Status_INFRA_FAILURE,
				SummaryMarkdown: fmt.Sprintf("fatal error in startup: %s", err),
			}
			bbClient.CloseAndDrain(ctx)
			os.Exit(1)
		}
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	opts := &host.Options{
		BaseBuild:      input.Build,
		ButlerLogLevel: logging.Warning,
		ViewerURL: fmt.Sprintf("https://%s/build/%d",
			input.Build.Infra.Buildbucket.Hostname, input.Build.Id),
	}
	opts.LogdogOutput, err = mkLogdogOutput(sctx, input.Build.Infra.Logdog)
	check(err)
	cwd, err := os.Getwd()
	check(errors.Annotate(err, "getting cwd").Err())

	opts.BaseDir = filepath.Join(cwd, "x")

	exePath, err := filepath.Abs(input.ExecutablePath)
	check(errors.Annotate(err, "absoluting exe path %q", input.ExecutablePath).Err())
	if runtime.GOOS == "windows" {
		exePath, err = resolveExe(exePath)
		check(errors.Annotate(err, "resolving %q", input.ExecutablePath).Err())
	}

	// TODO(iannucci): this is sketchy, but we preemptively add the log entries
	// for the top level user stdout/stderr streams.
	//
	// Really, `invoke.Start` is the one that knows how to arrange the
	// Output.Logs, but host.Run makes a copy of this build immediately. Find
	// a way to set these up nicely (maybe have opts.BaseBuild be a function
	// returning an immutable bbpb.Build?).
	input.Build.Output = &bbpb.Build_Output{
		Logs: []*bbpb.Log{
			{Name: "stdout", Url: "stdout"},
			{Name: "stderr", Url: "stderr"},
		},
	}

	initialJSONPB, err := (&jsonpb.Marshaler{
		OrigName: true, Indent: "  ",
	}).MarshalToString(input)
	check(errors.Annotate(err, "marshalling input args").Err())
	logging.Infof(ctx, "Input args:\n%s", initialJSONPB)

	builds, err := host.Run(cctx, opts, func(ctx context.Context) error {
		logging.Infof(ctx, "running luciexe: %q", exePath)
		logging.Infof(ctx, "  (cache dir): %q", input.CacheDir)
		subp, err := invoke.Start(ctx, exePath, input.Build, &invoke.Options{
			CacheDir: input.CacheDir,
		})
		if err != nil {
			return err
		}
		_, err = subp.Wait()
		return err
	})
	if err != nil {
		check(errors.Annotate(err, "could not start luciexe host environment").Err())
	}

	var finalStatus bbpb.Status

	// Now all we do is shuttle builds through to the buildbucket client channel
	// until there are no more builds to shuttle.
	for build := range builds {
		// TODO(iannucci): add backchannel from buildbucket prpc client to shut
		// down/cancel the build.
		bbClient.C <- build
		finalStatus = build.Status
	}

	if finalStatus != bbpb.Status_SUCCESS {
		return 1
	}
	return 0
}

func resolveExe(path string) (string, error) {
	if filepath.Ext(path) != "" {
		return path, nil
	}

	lme := errors.NewLazyMultiError(2)
	for i, ext := range []string{".exe", ".bat"} {
		candidate := path + ext
		if _, err := os.Stat(candidate); !lme.Assign(i, err) {
			return candidate, nil
		}
	}

	me := lme.Get().(errors.MultiError)
	return path, errors.Reason("cannot find .exe (%q) or .bat (%q)", me[0], me[1]).Err()
}
