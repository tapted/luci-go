// Copyright 2016 The LUCI Authors.
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

package buildbucket

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"

	"google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/auth/identity"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"
	"go.chromium.org/luci/buildbucket/protoutil"
	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"
	gitpb "go.chromium.org/luci/common/proto/git"
	"go.chromium.org/luci/common/sync/parallel"
	"go.chromium.org/luci/grpc/grpcutil"
	"go.chromium.org/luci/milo/api/config"
	"go.chromium.org/luci/milo/common"
	"go.chromium.org/luci/milo/common/model"
	"go.chromium.org/luci/milo/frontend/ui"
	"go.chromium.org/luci/server/auth"
	"go.chromium.org/luci/server/router"
)

var (
	ErrNotFound    = errors.Reason("Build not found").Tag(grpcutil.NotFoundTag).Err()
	ErrNotLoggedIn = errors.Reason("not logged in").Tag(grpcutil.UnauthenticatedTag).Err()
)

// BuildAddress constructs the build address of a buildbucketpb.Build.
// This is used as the key for the BuildSummary entity.
func BuildAddress(build *buildbucketpb.Build) string {
	if build == nil {
		return ""
	}
	num := strconv.FormatInt(build.Id, 10)
	if build.Number != 0 {
		num = strconv.FormatInt(int64(build.Number), 10)
	}
	b := build.Builder
	return fmt.Sprintf("luci.%s.%s/%s/%s", b.Project, b.Bucket, b.Builder, num)
}

// simplisticBlamelist returns a slice of ui.Commit for a build, and/or an error.
//
// HACK(iannucci) - Getting the frontend to render a proper blamelist will
// require some significant refactoring. To do this properly, we'll need:
//   * The frontend to get BuildSummary from the backend.
//   * BuildSummary to have a .PreviousBuild() API.
//   * The frontend to obtain the annotation streams itself (so it could see
//     the SourceManifest objects inside of them). Currently getRespBuild defers
//     to swarming's implementation of buildsource.ID.Get(), which only returns
//     the resp object.
func simplisticBlamelist(c context.Context, build *model.BuildSummary) (result []*ui.Commit, err error) {
	bs := build.GitilesCommit()
	if bs == nil {
		return
	}

	builds, commits, err := build.PreviousByGitilesCommit(c)
	switch {
	case err == nil || err == model.ErrUnknownPreviousBuild:
		// continue
	case status.Code(err) == codes.PermissionDenied:
		err = grpcutil.UnauthenticatedTag.Apply(err)
		return
	default:
		return
	}

	result = make([]*ui.Commit, 0, len(commits)+1)
	for _, commit := range commits {
		result = append(result, uiCommit(commit, protoutil.GitilesRepoURL(bs)))
	}
	logging.Infof(c, "Fetched %d commit blamelist from Gitiles", len(result))

	// this means that there were more than 100 commits in-between.
	if len(builds) == 0 && len(commits) > 0 {
		result = append(result, &ui.Commit{
			Description: "<blame list capped at 100 commits>",
			Revision:    &ui.Link{},
			AuthorName:  "<blame list capped at 100 commits>",
		})
	}

	return
}

func uiCommit(commit *gitpb.Commit, repoURL string) *ui.Commit {
	res := &ui.Commit{
		AuthorName:  commit.Author.Name,
		AuthorEmail: commit.Author.Email,
		Repo:        repoURL,
		Description: commit.Message,

		// TODO(iannucci): this use of links is very sloppy; the frontend should
		// know how to render a Commit without having Links embedded in it.
		Revision: ui.NewLink(
			commit.Id,
			repoURL+"/+/"+commit.Id, fmt.Sprintf("commit by %s", commit.Author.Email)),
	}
	res.CommitTime, _ = ptypes.Timestamp(commit.Committer.Time)
	res.File = make([]string, 0, len(commit.TreeDiff))
	for _, td := range commit.TreeDiff {
		// If file is moved, there is both new and old path,
		// from which we take only new path.
		// If a file is deleted, its new path is /dev/null.
		// In that case, we're only interested in the old path.
		switch {
		case td.NewPath != "" && td.NewPath != "/dev/null":
			res.File = append(res.File, td.NewPath)
		case td.OldPath != "":
			res.File = append(res.File, td.OldPath)
		}
	}
	return res
}

// GetBuildSummary fetches a build summary where the Context URI matches the
// given address.
func GetBuildSummary(c context.Context, id int64) (*model.BuildSummary, error) {
	// The host is set to prod because buildbot is hardcoded to talk to prod.
	uri := fmt.Sprintf("buildbucket://cr-buildbucket.appspot.com/build/%d", id)
	bs := make([]*model.BuildSummary, 0, 1)
	q := datastore.NewQuery("BuildSummary").Eq("ContextURI", uri).Limit(1)
	switch err := datastore.GetAll(c, q, &bs); {
	case err != nil:
		return nil, common.ReplaceNSEWith(err.(errors.MultiError), ErrNotFound)
	case len(bs) == 0:
		return nil, ErrNotFound
	default:
		return bs[0], nil
	}
}

// getBlame fetches blame information from Gitiles.
// This requires the BuildSummary to be indexed in Milo.
func getBlame(c context.Context, host string, b *buildbucketpb.Build) ([]*ui.Commit, error) {
	commit := b.GetInput().GetGitilesCommit()
	// No commit? No blamelist.
	if commit == nil {
		return nil, nil
	}
	// TODO(hinoka): This converts a buildbucketpb.Commit into a string
	// and back into a buildbucketpb.Commit.  That's a bit silly.
	return simplisticBlamelist(c, &model.BuildSummary{
		BuildKey:  MakeBuildKey(c, host, BuildAddress(b)),
		BuildSet:  []string{protoutil.GitilesBuildSet(commit)},
		BuilderID: LegacyBuilderIDString(b.Builder),
	})
}

// getBugLink attempts to formulate and return the build page bug link
// for the given build.
func getBugLink(c *router.Context, b *buildbucketpb.Build) (string, error) {
	project, err := common.GetProject(c.Context, b.Builder.GetProject())
	if err != nil || proto.Equal(&project.BuildBugTemplate, &config.BugTemplate{}) {
		return "", err
	}

	baseURL := "https://" + c.Request.Host
	builderPath := fmt.Sprintf("/p/%s/builders/%s/%s", b.Builder.GetProject(), b.Builder.GetBucket(), b.Builder.GetBuilder())

	buildURL, err := url.Parse(baseURL + builderPath + "/" + c.Params.ByName("numberOrId"))
	if err != nil {
		return "", errors.Annotate(err, "Unable to make build URL for build bug link.").Err()
	}

	builderURL, err := url.Parse(baseURL + builderPath)
	if err != nil {
		return "", errors.Annotate(err, "Unable to make builder URL for build bug link.").Err()
	}

	return MakeBuildBugLink(&project.BuildBugTemplate, map[string]interface{}{
		"Build":          b,
		"MiloBuildUrl":   buildURL,
		"MiloBuilderUrl": builderURL,
	})
}

// searchBuildset creates a searchBuildsRequest that looks for a buildset tag.
func searchBuildset(buildset string, fields *field_mask.FieldMask) *buildbucketpb.SearchBuildsRequest {
	return &buildbucketpb.SearchBuildsRequest{
		Predicate: &buildbucketpb.BuildPredicate{
			Tags: []*buildbucketpb.StringPair{{Key: "buildset", Value: buildset}},
		},
		Fields:   fields,
		PageSize: 1000,
	}
}

var summaryBuildsMask = &field_mask.FieldMask{
	Paths: []string{
		"builds.*.id",
		"builds.*.builder",
		"builds.*.number",
		"builds.*.create_time",
		"builds.*.start_time",
		"builds.*.end_time",
		"builds.*.update_time",
		"builds.*.status",
		"builds.*.summary_markdown",
	},
}

// getRelatedBuilds fetches build summaries of builds with the same buildset as b.
func getRelatedBuilds(c context.Context, now *timestamp.Timestamp, client buildbucketpb.BuildsClient, b *buildbucketpb.Build) ([]*ui.Build, error) {
	var bs []string
	for _, buildset := range protoutil.BuildSets(b) {
		// HACK(hinoka): Remove the commit/git/ buildsets because we know they're redundant
		// with the commit/gitiles/ buildsets, and we don't need to ask Buildbucket twice.
		if strings.HasPrefix(buildset, "commit/git/") {
			continue
		}
		bs = append(bs, buildset)
	}
	if len(bs) == 0 {
		// No buildset? No builds.
		return nil, nil
	}

	// Do the search request.
	// Use multiple requests instead of a single batch request.
	// A single large request is CPU bound to a single GAE instance on the buildbucket side.
	// Multiple requests allows the use of multiple GAE instances, therefore more parallelism.
	resps := make([]*buildbucketpb.SearchBuildsResponse, len(bs))
	if err := parallel.WorkPool(8, func(ch chan<- func() error) {
		for i, buildset := range bs {
			i := i
			buildset := buildset
			ch <- func() (err error) {
				logging.Debugf(c, "Searching for %s (%d)", buildset, i)
				resps[i], err = client.SearchBuilds(c, searchBuildset(buildset, summaryBuildsMask))
				return
			}
		}
	}); err != nil {
		return nil, err
	}

	// Dedupe builds.
	// It's possible since we've made multiple requests that we got back the same builds
	// multiple times.
	seen := map[int64]bool{} // set of build IDs.
	result := []*ui.Build{}
	for _, resp := range resps {
		for _, rb := range resp.GetBuilds() {
			if seen[rb.Id] {
				continue
			}
			seen[rb.Id] = true
			result = append(result, &ui.Build{
				Build: rb,
				Now:   now,
			})
		}
	}

	// Sort builds by ID.
	sort.Slice(result, func(i, j int) bool { return result[i].Id < result[j].Id })

	return result, nil
}

var builderIDMask = &field_mask.FieldMask{
	Paths: []string{
		"builder",
		"number",
	},
}

// GetBuilderID returns the builder, and maybe the build number, for a build id.
func GetBuilderID(c context.Context, id int64) (builder *buildbucketpb.BuilderID, number int32, err error) {
	client, err := getBuildbucketClient(c)
	if err != nil {
		return
	}
	br, err := client.GetBuild(c, &buildbucketpb.GetBuildRequest{
		Id:     id,
		Fields: builderIDMask,
	})
	switch grpcutil.Code(err) {
	case codes.OK:
		builder = br.Builder
		number = br.Number
	case codes.NotFound:
		if auth.CurrentIdentity(c) == identity.AnonymousIdentity {
			err = ErrNotLoggedIn
			return
		}
		fallthrough
	case codes.PermissionDenied:
		err = ErrNotFound
	}
	return
}

var (
	fullBuildMask = &field_mask.FieldMask{
		Paths: []string{
			"id",
			"builder",
			"number",
			"created_by",
			"canceled_by",
			"create_time",
			"start_time",
			"end_time",
			"update_time",
			"status",
			"input",
			"output",
			"steps",
			"infra",
			"tags",
			"summary_markdown",
		},
	}
	tagsAndGitilesMask = &field_mask.FieldMask{
		Paths: []string{
			"id",
			"number",
			"builder",
			"input.gitiles_commit",
			"tags",
		},
	}
)

// GetBuildPage fetches the full set of information for a Milo build page from Buildbucket.
// Including the blamelist and other auxiliary information.
func GetBuildPage(ctx *router.Context, br buildbucketpb.GetBuildRequest, forceBlamelist bool) (*ui.BuildPage, error) {
	now, _ := ptypes.TimestampProto(clock.Now(ctx.Context))

	c := ctx.Context
	host, err := getHost(c)
	if err != nil {
		return nil, err
	}
	client, err := buildbucketClient(c, host, auth.AsUser)
	if err != nil {
		return nil, err
	}

	var b *buildbucketpb.Build
	var blame []*ui.Commit
	var blameErr error
	if err = parallel.FanOutIn(func(ch chan<- func() error) {
		ch <- func() (err error) {
			fullbr := br // Copy request
			fullbr.Fields = fullBuildMask
			b, err = client.GetBuild(c, &fullbr)
			return common.TagGRPC(c, err)
		}

		// Fetch a small build with just a tiny bit of information.
		// We use this to get the Gitiles tag so that we can fetch
		// related builds and blamelist in parallel.
		smallbr := br // Copy request
		smallbr.Fields = tagsAndGitilesMask
		sb, err := client.GetBuild(c, &smallbr)
		if err != nil {
			return
		}
		ch <- func() error {
			timeout := 1 * time.Second
			if forceBlamelist {
				timeout = 55 * time.Second
			}
			nc, cancel := context.WithTimeout(c, timeout)
			defer cancel()
			blame, blameErr = getBlame(nc, host, sb)
			return nil
		}
	}); err != nil {
		return nil, err
	}
	link, err := getBugLink(ctx, b)
	logging.Infof(c, "Got all the things")
	return &ui.BuildPage{
		Build: ui.Build{
			Build: b,
			Now:   now,
		},
		Blame:           blame,
		BuildBugLink:    link,
		BuildbucketHost: host,
		BlamelistError:  blameErr,
		ForcedBlamelist: forceBlamelist,
	}, err
}

// GetRelatedBuildsTable fetches all the related builds of the given build from Buildbucket.
func GetRelatedBuildsTable(c context.Context, buildbucketID int64) (*ui.RelatedBuildsTable, error) {
	now, _ := ptypes.TimestampProto(clock.Now(c))

	client, err := getBuildbucketClient(c)
	if err != nil {
		return nil, err
	}

	build, err := client.GetBuild(c, &buildbucketpb.GetBuildRequest{
		Id:     buildbucketID,
		Fields: tagsAndGitilesMask,
	})
	if err != nil {
		return nil, err
	}

	relatedBuilds, err := getRelatedBuilds(c, now, client, build)
	if err != nil {
		return nil, err
	}

	return &ui.RelatedBuildsTable{
		Build: ui.Build{
			Build: build,
			Now:   now,
		},
		RelatedBuilds: relatedBuilds,
	}, nil
}

// CancelBuild cancels the build with the given ID.
func CancelBuild(c context.Context, id int64, reason string) (*buildbucketpb.Build, error) {
	client, err := getBuildbucketClient(c)
	if err != nil {
		return nil, err
	}

	return client.CancelBuild(c, &buildbucketpb.CancelBuildRequest{
		Id:              id,
		SummaryMarkdown: reason,
	})
}

// RetryBuild retries the build with the given ID and returns the new build.
func RetryBuild(c context.Context, buildbucketID int64, requestID string) (*buildbucketpb.Build, error) {
	client, err := getBuildbucketClient(c)
	if err != nil {
		return nil, err
	}

	return client.ScheduleBuild(c, &buildbucketpb.ScheduleBuildRequest{
		RequestId:       requestID,
		TemplateBuildId: buildbucketID,
	})
}

func getBuildbucketClient(c context.Context) (buildbucketpb.BuildsClient, error) {
	host, err := getHost(c)
	if err != nil {
		return nil, err
	}
	client, err := buildbucketClient(c, host, auth.AsUser)
	if err != nil {
		return nil, err
	}
	return client, nil
}
