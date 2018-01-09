// Copyright 2017 The LUCI Authors.
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

package gitiles

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"

	"go.chromium.org/luci/common/proto/git"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestTimestamp(t *testing.T) {
	t.Parallel()

	Convey("Marshal and Unmarshal ts", t, func() {
		// Nanoseconds must be zero because the string format in between
		// does not contain nanoseconds.
		tBefore := ts{time.Date(12, 2, 5, 6, 1, 3, 0, time.UTC)}
		bytes, err := json.Marshal(tBefore)
		So(err, ShouldBeNil)

		var tAfter ts
		err = json.Unmarshal(bytes, &tAfter)
		So(err, ShouldBeNil)

		So(tBefore, ShouldResemble, tAfter)
	})
}

func TestUser(t *testing.T) {
	t.Parallel()

	Convey(`Test user.Proto`, t, func() {
		u := &user{
			Name:  "Some name",
			Email: "some.name@example.com",
			Time:  ts{time.Date(2016, 3, 9, 3, 46, 18, 0, time.UTC)},
		}

		Convey(`basic`, func() {
			uPB, err := u.Proto()
			So(err, ShouldBeNil)
			So(uPB, ShouldResemble, &git.Commit_User{
				Name:  "Some name",
				Email: "some.name@example.com",
				Time: &timestamp.Timestamp{
					Seconds: 1457495178,
				},
			})
		})

		Convey(`empty ts`, func() {
			u.Time = ts{}
			uPB, err := u.Proto()
			So(err, ShouldBeNil)
			So(uPB, ShouldResemble, &git.Commit_User{
				Name:  "Some name",
				Email: "some.name@example.com",
			})
		})
	})
}

func TestTreeDiff(t *testing.T) {
	t.Parallel()

	Convey(`Test treeDiff.Proto`, t, func() {
		td := &treeDiff{
			Type:    "MODIFY",
			OldID:   strings.Repeat("deadbeef", 5),
			OldPath: "some/path",
			OldMode: 0666,
			NewID:   strings.Repeat("daff0d11", 5),
			NewPath: "some/path",
			NewMode: 0666,
		}

		Convey(`basic`, func() {
			tdPD, err := td.Proto()
			So(err, ShouldBeNil)
			So(tdPD, ShouldResemble, &git.Commit_TreeDiff{
				Type:    git.Commit_TreeDiff_MODIFY,
				OldId:   bytes.Repeat([]byte{0xde, 0xad, 0xbe, 0xef}, 5),
				OldPath: "some/path",
				OldMode: 0666,
				NewId:   bytes.Repeat([]byte{0xda, 0xff, 0x0d, 0x11}, 5),
				NewPath: "some/path",
				NewMode: 0666,
			})
		})

		Convey(`bad type`, func() {
			td.Type = "Meep"
			_, err := td.Proto()
			So(err, ShouldErrLike, "bad change type")
		})

		Convey(`bad OldID`, func() {
			td.OldID = "Meep"
			_, err := td.Proto()
			So(err, ShouldErrLike, "decoding OldID")
		})

		Convey(`bad NewID`, func() {
			td.NewID = "Meep"
			_, err := td.Proto()
			So(err, ShouldErrLike, "decoding NewID")
		})
	})
}

func TestCommit(t *testing.T) {
	t.Parallel()

	Convey(`Test commit.Proto`, t, func() {
		c := &commit{
			Commit: strings.Repeat("deadbeef", 5),
			Tree:   strings.Repeat("ac1df00d", 5),
			Parents: []string{
				strings.Repeat("d15c0bee", 5),
				strings.Repeat("daff0d11", 5),
			},
			Author:    user{"author", "author@example.com", ts{time.Date(2016, 3, 9, 3, 46, 18, 0, time.UTC)}},
			Committer: user{"committer", "committer@example.com", ts{time.Date(2016, 3, 9, 3, 46, 18, 0, time.UTC)}},
			Message:   "I am\na\nbanana",
		}

		Convey(`basic`, func() {
			cPB, err := c.Proto()
			So(err, ShouldBeNil)
			So(cPB, ShouldResemble, &git.Commit{
				Id:   bytes.Repeat([]byte{0xde, 0xad, 0xbe, 0xef}, 5),
				Tree: bytes.Repeat([]byte{0xac, 0x1d, 0xf0, 0x0d}, 5),
				Parents: [][]byte{
					bytes.Repeat([]byte{0xd1, 0x5c, 0x0b, 0xee}, 5),
					bytes.Repeat([]byte{0xda, 0xff, 0x0d, 0x11}, 5),
				},
				Author: &git.Commit_User{
					Name:  "author",
					Email: "author@example.com",
					Time:  &timestamp.Timestamp{Seconds: 1457495178},
				},
				Committer: &git.Commit_User{
					Name:  "committer",
					Email: "committer@example.com",
					Time:  &timestamp.Timestamp{Seconds: 1457495178},
				},
				Message: "I am\na\nbanana",
			})
		})

		Convey(`bad id`, func() {
			c.Commit = "nerp"
			_, err := c.Proto()
			So(err, ShouldErrLike, "decoding id")
		})

		Convey(`bad tree`, func() {
			c.Tree = "nerp"
			_, err := c.Proto()
			So(err, ShouldErrLike, "decoding tree")
		})

		Convey(`bad parent`, func() {
			c.Parents[0] = "nerp"
			_, err := c.Proto()
			So(err, ShouldErrLike, "decoding parent 0")
		})
	})
}