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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/context"

	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/rpc/v1"
	typepb "go.chromium.org/luci/resultdb/proto/type"
)

const (
	testInstantiationKey = "param/instantiation"
	testParameterKey     = "param/id"
)

var (
	// Prefixes that may be present in the test name and must be stripped before forming the base path.
	prefixes = []string{"MANUAL_", "PRE_"}

	// Java base paths aren't actually GTest but use the same launcher output format.
	javaPathRE = regexp.MustCompile(`^[\w.]+#[\w]+$`)

	// Test base paths look like FooTest.DoesBar: "FooTest" is the suite and "DoesBar" the test name.
	basePathRE = regexp.MustCompile(`^(\w+)\.(\w+)$`)

	// Type parametrized test examples:
	// - MyInstantiation/FooTest/1.DoesBar
	// - FooTest/1.DoesBar
	// - FooType/MyType.DoesBar
	//
	// In the above examples, "FooTest" is the suite, "DoesBar" the test name, "MyInstantiation" the
	// optional instantiation, "1" the index of the type on which the test has been instantiated, if
	// no string representation for the type has been provided, and "MyType" is the user-provided
	// string representation of the type on which the test has been instantiated.
	typeParamRE = regexp.MustCompile(`^((\w+)/)?(\w+)/(\w+)\.(\w+)$`)

	// Value parametrized tests examples:
	// - MyInstantiation/FooTest.DoesBar/1
	// - FooTest.DoesBar/1
	// - FooTest.DoesBar/TestValue
	//
	// In the above examples, "FooTest" is the suite, "DoesBar" the test name, "MyInstantiation" the
	// optional instantiation, "1" the index of the value on which the test has been instantiated, if
	// no string representation for the value has been provided, and "TestValue" is the user-provided
	// string representation of the value on which the test has been instantiated.
	valueParamRE = regexp.MustCompile(`^((\w+)/)?(\w+)\.(\w+)/(\w+)$`)
)

// GTestResults represents the structure as described to be generated in
// https://cs.chromium.org/chromium/src/base/test/launcher/test_results_tracker.h?l=83&rcl=96020cfd447cb285acfa1a96c37a67ed22fa2499
// (base::TestResultsTracker::SaveSummaryAsJSON)
//
// Fields not used by Test Results are omitted.
type GTestResults struct {
	AllTests   []string `json:"all_tests"`
	GlobalTags []string `json:"global_tags"`

	// PerIterationData is a vector of run iterations, each mapping test names to a list of test data.
	PerIterationData []map[string][]*GTestRunResult `json:"per_iteration_data"`

	// TestLocations maps test names to their location in code.
	TestLocations map[string]*Location `json:"test_locations"`
}

// GTestRunResult represents the per_iteration_data as described in
// https://cs.chromium.org/chromium/src/base/test/launcher/test_results_tracker.h?l=83&rcl=96020cfd447cb285acfa1a96c37a67ed22fa2499
// (base::TestResultsTracker::SaveSummaryAsJSON)
//
// Fields not used by Test Results are omitted.
type GTestRunResult struct {
	Status        string `json:"status"`
	ElapsedTimeMs int    `json:"elapsed_time_ms"`

	LosslessSnippet     bool   `json:"losless_snippet"`
	OutputSnippetBase64 string `json:"output_snippet_base64"`
}

// Location describes a code location.
type Location struct {
	File string `json:"file"`
	Line int    `json:"line"`
}

// ConvertFromJSON reads the provided reader into the receiver.
//
// The receiver is cleared and its fields overwritten.
func (r *GTestResults) ConvertFromJSON(ctx context.Context, reader io.Reader) error {
	*r = GTestResults{}
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return err
	}

	if len(r.AllTests) == 0 {
		return errors.Reason(`missing "all_tests" field in JSON`).Err()
	}

	return nil
}

// ToProtos converts test results in r []*pb.TestResult and updates inv
// in-place accordingly.
// If an error is returned, inv is left unchanged.
//
// Does not populate TestResult.Name.
func (r *GTestResults) ToProtos(ctx context.Context, testPathPrefix string, inv *pb.Invocation) ([]*pb.TestResult, error) {
	// In theory, we can have multiple iterations. This seems rare in practice, so log if we do see
	// more than one to confirm and track.
	if len(r.PerIterationData) > 1 {
		logging.Infof(ctx, "Got %d GTest iterations", len(r.PerIterationData))
	}

	// Assume the invocation was not interrupted; if any results are NOTRUN,
	// we'll mark as otherwise.
	interrupted := false

	var ret []*pb.TestResult
	var testNames []string
	for _, data := range r.PerIterationData {
		// Sort the test name to make the output deterministic.
		testNames = testNames[:0]
		for name := range data {
			testNames = append(testNames, name)
		}
		sort.Strings(testNames)

		for _, name := range testNames {
			baseName, params, err := extractGTestParameters(name)
			if err != nil {
				return nil, errors.Annotate(err,
					"failed to extract test base name and parameters from %q", name).Err()
			}

			testPath := testPathPrefix + baseName

			for i, result := range data[name] {
				// Store the processed test result into the correct part of the overall map.
				rpb, err := r.convertTestResult(ctx, testPath, name, result)
				if err != nil {
					return nil, errors.Annotate(err,
						"iteration %d of test %s failed to convert run result", i, name).Err()
				}

				if len(params) > 0 {
					rpb.Variant = &typepb.Variant{Def: params}
				}

				// TODO(jchinlee): Verify that it's indeed the case that getting NOTRUN results in the final
				// results indicates the task was incomplete.
				// TODO(jchinlee): Check how unexpected SKIPPED tests should be handled.
				if result.Status == "NOTRUN" {
					interrupted = true
				}

				ret = append(ret, rpb)
			}
		}
	}

	// The code below does not return errors, so it is safe to make in-place
	// modifications of inv.

	if interrupted {
		inv.State = pb.Invocation_INTERRUPTED
	}

	// Populate the tags.
	for _, tag := range r.GlobalTags {
		inv.Tags = append(inv.Tags, pbutil.StringPair("gtest_global_tag", tag))
	}
	inv.Tags = append(inv.Tags, pbutil.StringPair(OriginalFormatTagKey, FormatGTest))

	pbutil.NormalizeInvocation(inv)
	return ret, nil
}

func fromGTestStatus(s string) (pb.TestStatus, error) {
	switch s {
	case "SUCCESS":
		return pb.TestStatus_PASS, nil
	case "FAILURE":
		return pb.TestStatus_FAIL, nil
	case "FAILURE_ON_EXIT":
		return pb.TestStatus_FAIL, nil
	case "TIMEOUT":
		return pb.TestStatus_ABORT, nil
	case "CRASH":
		return pb.TestStatus_CRASH, nil
	case "SKIPPED":
		return pb.TestStatus_SKIP, nil
	case "EXCESSIVE_OUTPUT":
		return pb.TestStatus_FAIL, nil
	case "NOTRUN":
		return pb.TestStatus_SKIP, nil
	default:
		// This would only happen if the set of possible GTest result statuses change and resultsdb has
		// not been updated to match.
		return pb.TestStatus_STATUS_UNSPECIFIED, errors.Reason("unknown GTest status %q", s).Err()
	}
}

// extractGTestParameters extracts parameters from a test path as a mapping with "param/" keys.
func extractGTestParameters(testPath string) (basePath string, params map[string]string, err error) {
	var suite, name string

	// If this is a JUnit tests, don't try to extract parameters.
	// TODO: investigate handling parameters for JUnit tests.
	if match := javaPathRE.FindStringSubmatch(testPath); match != nil {
		basePath = testPath
		return
	}

	// Tests can be only one of type- or value-parametrized, if parametrized at all.
	params = map[string]string{}
	if match := typeParamRE.FindStringSubmatch(testPath); match != nil {
		// Extract type parameter.
		suite = match[3]
		name = match[5]
		params[testInstantiationKey] = match[2]
		params[testParameterKey] = match[4]
	} else if match := valueParamRE.FindStringSubmatch(testPath); match != nil {
		// Extract value parameter.
		suite = match[3]
		name = match[4]
		params[testInstantiationKey] = match[2]
		params[testParameterKey] = match[5]
	} else if match := basePathRE.FindStringSubmatch(testPath); match != nil {
		// Otherwise our testPath should not be parametrized, so extract the suite and name.
		suite = match[1]
		name = match[2]
	} else {
		// Otherwise testPath format is unrecognized.
		err = errors.Reason("test path of unknown format").Err()
		return
	}

	// Strip prefixes from test name if necessary.
	for {
		strippedName := name
		for _, prefix := range prefixes {
			strippedName = strings.TrimPrefix(strippedName, prefix)
		}
		if strippedName == name {
			break
		}
		name = strippedName
	}
	basePath = fmt.Sprintf("%s.%s", suite, name)

	return
}

func (r *GTestResults) convertTestResult(ctx context.Context, testPath, name string, result *GTestRunResult) (*pb.TestResult, error) {
	status, err := fromGTestStatus(result.Status)
	if err != nil {
		return nil, err
	}

	rpb := &pb.TestResult{
		TestPath: testPath,
		Expected: status == pb.TestStatus_PASS,
		Status:   status,
		Tags: pbutil.StringPairs(
			// Store the original GTest status.
			"gtest_status", result.Status,
			// Store the correct output snippet.
			"lossless_snippet", strconv.FormatBool(result.LosslessSnippet),
		),
	}

	// Do not set duration if it is unknown.
	if result.ElapsedTimeMs != 0 {
		rpb.Duration = secondsToDuration(1e-6 * float64(result.ElapsedTimeMs))
	}

	// Write the summary.
	if result.OutputSnippetBase64 != "" {
		outputBytes, err := base64.StdEncoding.DecodeString(result.OutputSnippetBase64)
		if err != nil {
			// Log the error, but we shouldn't fail to convert an entire invocation just because we can't
			// convert a summary.
			logging.Errorf(ctx, "Failed to convert OutputSnippetBase64 %q", result.OutputSnippetBase64)
		} else {
			rpb.OutputArtifacts = append(rpb.OutputArtifacts, &pb.Artifact{
				Name:        "gtest_snippet.txt",
				ContentType: "text/plain",
				Contents:    outputBytes,
			})
		}
	}

	// Store the test code location.
	if loc, ok := r.TestLocations[name]; ok {
		rpb.Tags = append(rpb.Tags,
			pbutil.StringPair("gtest_file", loc.File),
			pbutil.StringPair("gtest_line", strconv.Itoa(loc.Line)),
		)
	}

	return rpb, nil
}
