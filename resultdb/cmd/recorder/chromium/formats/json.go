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

package formats

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"go.chromium.org/luci/common/data/stringset"
	"go.chromium.org/luci/common/data/text/indented"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/rpc/v1"
)

const testNamePrefixKey = "test_name_prefix"

var testRunSubdirRe = regexp.MustCompile("/retry_([0-9]+)/")

// JSONTestResults represents the structure in
// https://chromium.googlesource.com/chromium/src/+/master/docs/testing/json_test_results_format.md
//
// Deprecated fields and fields not used by Test Results are omitted.
type JSONTestResults struct {
	Interrupted bool `json:"interrupted"`

	PathDelimiter string `json:"path_delimiter"`

	TestsRaw json.RawMessage `json:"tests"`
	Tests    map[string]*TestFields

	Version int32 `json:"version"`

	ArtifactTypes map[string]string `json:"artifact_types"`

	BuildNumber string `json:"build_number"`
	BuilderName string `json:"builder_name"`

	// Metadata associated with results, which may include a list of expectation_files, or
	// test_name_prefix e.g. in GPU tests (distinct from test_path_prefix passed in the recorder API
	// request).
	Metadata map[string]json.RawMessage `json:"metadata"`
}

// TestFields represents the test fields structure in
// https://chromium.googlesource.com/chromium/src/+/master/docs/testing/json_test_results_format.md
//
// Deprecated fields and fields not used by Test Results are omitted.
type TestFields struct {
	Actual   string `json:"actual"`
	Expected string `json:"expected"`

	Artifacts map[string][]string `json:"artifacts"`

	Time  float64   `json:"time"`
	Times []float64 `json:"times"`
}

// ConvertFromJSON converts a JSON of test results in the JSON Test Results
// format to the internal struct format.
//
// The receiver is cleared and its fields overwritten.
func (r *JSONTestResults) ConvertFromJSON(ctx context.Context, reader io.Reader) error {
	*r = JSONTestResults{}
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return err
	}

	// Convert Tests and return.
	if err := r.convertTests("", r.TestsRaw); err != nil {
		return err
	}
	return nil
}

// ToProtos converts test results in r []*pb.TestResult and updates inv
// in-place accordingly.
// If an error is returned, inv is left unchanged.
//
// Takes outputsToProcess, the isolated outputs associated with the task, to use to populate
// artifacts, and deletes any that are successfully processed.
// Does not populate TestResult.Name; that happens server-side on RPC response.
func (r *JSONTestResults) ToProtos(ctx context.Context, testPathPrefix string, inv *pb.Invocation, outputsToProcess map[string]*pb.Artifact) ([]*pb.TestResult, error) {
	if r.Version != 3 {
		return nil, errors.Reason("unknown JSON Test Results version %d", r.Version).Err()
	}

	// Sort the test name to make the output deterministic.
	testNames := make([]string, 0, len(r.Tests))
	for name := range r.Tests {
		testNames = append(testNames, name)
	}
	sort.Strings(testNames)

	ret := make([]*pb.TestResult, 0, len(r.Tests))
	for _, name := range testNames {
		testPath := testPathPrefix + name

		// Populate protos.
		unresolvedOutputs, err := r.Tests[name].toProtos(ctx, &ret, testPath, outputsToProcess)
		if err != nil {
			return nil, errors.Annotate(err, "test %q failed to convert run fields", name).Err()
		}

		// If any outputs cannot be processed, don't cause the rest of processing to fail, but do log.
		if len(unresolvedOutputs) > 0 {
			logging.Errorf(ctx,
				"Test %s could not generate artifact protos for the following:\n%s",
				testPath,
				artifactsToString(unresolvedOutputs))
		}
	}

	// Get tags from metadata if any.
	tags, err := r.extractTags()
	if err != nil {
		return nil, err
	}

	// The code below does not return errors, so it is safe to make in-place
	// modifications of inv.

	if r.Interrupted {
		inv.State = pb.Invocation_INTERRUPTED
	}

	inv.Tags = append(inv.Tags, pbutil.StringPair(OriginalFormatTagKey, FormatJTR))
	for _, tag := range tags {
		inv.Tags = append(inv.Tags, pbutil.StringPair("json_format_tag", tag))
	}
	if r.BuildNumber != "" {
		inv.Tags = append(inv.Tags, pbutil.StringPair("build_number", r.BuildNumber))
	}

	pbutil.NormalizeInvocation(inv)
	return ret, nil
}

// convertTests converts the trie of tests.
func (r *JSONTestResults) convertTests(curPath string, curNode json.RawMessage) error {
	// curNode should certainly be a map.
	var maybeNode map[string]json.RawMessage
	if err := json.Unmarshal(curNode, &maybeNode); err != nil {
		return errors.Annotate(err, "%q not map[string]json.RawMessage", curNode).Err()
	}

	// Convert the tree.
	for key, value := range maybeNode {
		// Set up test path.
		delim := "/"
		testPath := key
		if r.PathDelimiter != "" {
			delim = r.PathDelimiter
		}

		if curPath != "" {
			testPath = fmt.Sprintf("%s%s%s", curPath, delim, key)
		} else {
			if prefixJSON, ok := r.Metadata[testNamePrefixKey]; ok {
				var prefix string
				if err := json.Unmarshal(prefixJSON, &prefix); err != nil {
					return errors.Annotate(err, "%s not string, got %q", testNamePrefixKey, prefixJSON).Err()
				}
				testPath = prefix + key
			}
		}

		// Try to unmarshal value to TestFields. We check success by checking fields we expect to
		// be populated.
		maybeFields := &TestFields{}
		json.Unmarshal(value, maybeFields)
		if maybeFields.Actual != "" && maybeFields.Expected != "" {
			if r.Tests == nil {
				r.Tests = make(map[string]*TestFields)
			}
			r.Tests[testPath] = maybeFields
			continue
		}

		// Otherwise, try to process it as an intermediate node.
		if err := r.convertTests(testPath, value); err != nil {
			return errors.Annotate(err, "error attempting conversion of %q as intermediated node", value).Err()
		}
	}
	return nil
}

// extractTags tries to read the optional "tags" field in "metadata" as a slice of strings.
func (r *JSONTestResults) extractTags() ([]string, error) {
	maybeTags, ok := r.Metadata["tags"]
	if !ok {
		return nil, nil
	}

	var tags []string
	if err := json.Unmarshal(maybeTags, &tags); err != nil {
		return nil, errors.Annotate(err, "tags not []string, got %q", maybeTags).Err()
	}

	return tags, nil
}

func fromJSONStatus(s string) (pb.TestStatus, error) {
	switch s {
	case "CRASH":
		return pb.TestStatus_CRASH, nil
	case "FAIL":
		return pb.TestStatus_FAIL, nil
	case "PASS":
		return pb.TestStatus_PASS, nil
	case "SKIP":
		return pb.TestStatus_SKIP, nil
	case "TIMEOUT":
		return pb.TestStatus_ABORT, nil

	// The below are web test-specific statuses. They are officially deprecated, but in practice
	// still generated by the tests and should be converted.
	case "IMAGE", "TEXT", "IMAGE+TEXT", "AUDIO", "LEAK", "MISSING":
		return pb.TestStatus_FAIL, nil

	default:
		// There are a number of web test-specific statuses not handled here as they are deprecated.
		return 0, errors.Reason("unknown or unexpected JSON Test Format status %s", s).Err()
	}
}

// testArtifactsPerRun maps a run index to a map of run index to slice of
// associated *pb.Artifacts.
type testArtifactsPerRun map[int][]*pb.Artifact

// toProtos converts the TestFields into zero or more pb.TestResult and
// appends them to dest.
//
// Any artifacts that could not be processed are returned.
// TODO(jchinlee): once we've curated the artifacts to process, make unprocessed artifacts error.
func (f *TestFields) toProtos(ctx context.Context, dest *[]*pb.TestResult, testPath string, outputsToProcess map[string]*pb.Artifact) (map[string][]string, error) {
	// Process statuses.
	actualStatuses := strings.Split(f.Actual, " ")
	expectedSet := stringset.NewFromSlice(strings.Split(f.Expected, " ")...)

	// Process times.
	// Time and Times are both optional, but if Times is present, its length should match the number
	// of runs. Otherwise we have only Time as the duration of the first run.
	if len(f.Times) > 0 && len(f.Times) != len(actualStatuses) {
		return nil, errors.Reason(
			"%d durations populated but has %d test statuses; should match",
			len(f.Times), len(actualStatuses)).Err()
	}

	var durations []float64
	if len(f.Times) > 0 {
		durations = f.Times
	} else if f.Time != 0 { // Do not set duration if it is unknown.
		durations = []float64{f.Time}
	}

	// Get artifacts.
	// We expect that if we have any artifacts, the number of runs from deriving the artifacts
	// should match the number of actual runs. Because the arts are a map from run index to
	// *pb.Artifacts slice, we will not error if artifacts are missing for a run, but log a warning
	// in case the number of runs do not match each other for further investigation.
	arts, unresolved := f.getArtifacts(outputsToProcess)
	if len(arts) > 0 && len(actualStatuses) != len(arts) {
		logging.Warningf(ctx,
			"Number of runs of test %s (%d) does not match number of runs generated from artifacts (%d)",
			len(actualStatuses), len(arts), testPath)
	}

	// Populate protos.
	for i, runStatus := range actualStatuses {
		status, err := fromJSONStatus(runStatus)
		if err != nil {
			return nil, err
		}

		tr := &pb.TestResult{
			TestPath:        testPath,
			Expected:        expectedSet.Has(runStatus),
			Status:          status,
			Tags:            pbutil.StringPairs("json_format_status", runStatus),
			OutputArtifacts: arts[i],
		}

		if i < len(durations) {
			tr.Duration = secondsToDuration(durations[i])
		}

		pbutil.NormalizeTestResult(tr)
		*dest = append(*dest, tr)
	}

	return unresolved, nil
}

// getArtifacts gets pb.Artifacts corresponding to the TestField's artifacts.
//
// It tries to derive the pb.Artifacts in the following order:
//   - look for them in the isolated outputs represented as pb.Artifacts
//   - check if they're a known special case
//   - fail to process and mark them as `unresolvedArtifacts`
func (f *TestFields) getArtifacts(outputsToProcess map[string]*pb.Artifact) (artifacts testArtifactsPerRun, unresolvedArtifacts map[string][]string) {
	artifacts = testArtifactsPerRun{}
	unresolvedArtifacts = map[string][]string{}

	for name, paths := range f.Artifacts {
		for i, path := range paths {
			// Get the run ID of the artifact. Defaults to 0 (i.e. assumes there is only one run).
			runID, err := artifactRunID(path)
			if err != nil {
				unresolvedArtifacts[name] = append(unresolvedArtifacts[name], path)
				continue
			}

			// Look for the path in isolated outputs.
			if art, ok := outputsToProcess[path]; ok {
				artifacts[runID] = append(artifacts[runID], art)
				delete(outputsToProcess, path)
				continue
			}

			// If the name is otherwise understood by ResultDB, process it.
			// So far, that's only gold_triage_links.
			if name == "gold_triage_link" || name == "triage_link_for_entire_cl" {
				// We don't expect more than one triage link per test run, but if there is more than one,
				// suffix the name with index to ensure we retain it too.
				if i > 0 {
					name = fmt.Sprintf("%s_%d", name, i)
				}

				artifacts[runID] = append(artifacts[runID], &pb.Artifact{Name: name, ViewUrl: path})
				continue
			}

			// Otherwise, could not populate artifact, so mark it as unresolved.
			unresolvedArtifacts[name] = append(unresolvedArtifacts[name], path)
		}
	}

	return
}

// artifactRunID extracts a run ID, defaulting to 0, or error if it doesn't recognize the format.
func artifactRunID(path string) (int, error) {
	if m := testRunSubdirRe.FindStringSubmatch(path); m != nil {
		return strconv.Atoi(m[1])
	}

	// No retry_<i> subdirectory, so assume it's the first/0th run.
	return 0, nil
}

// artifactsToString converts the given name->paths artifacts map to a string for logging.
func artifactsToString(arts map[string][]string) string {
	names := make([]string, 0, len(arts))
	for name := range arts {
		names = append(names, name)
	}
	sort.Strings(names)

	var msg bytes.Buffer
	w := &indented.Writer{Writer: &msg}
	for _, name := range names {
		fmt.Fprintln(w, name)
		w.Level++
		for _, p := range arts[name] {
			fmt.Fprintln(w, p)
		}
		w.Level--
	}
	return msg.String()
}
