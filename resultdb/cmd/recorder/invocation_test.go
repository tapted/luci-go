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
	"context"
	"testing"
	"time"

	"cloud.google.com/go/spanner"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"go.chromium.org/luci/common/clock"
	"go.chromium.org/luci/common/clock/testclock"
	"go.chromium.org/luci/grpc/grpcutil"

	"go.chromium.org/luci/resultdb/internal/span"
	"go.chromium.org/luci/resultdb/internal/testutil"
	"go.chromium.org/luci/resultdb/pbutil"
	pb "go.chromium.org/luci/resultdb/proto/rpc/v1"

	. "github.com/smartystreets/goconvey/convey"
	. "go.chromium.org/luci/common/testing/assertions"
)

func TestMutateInvocation(t *testing.T) {
	Convey("MayMutateInvocation", t, func() {
		ctx := testutil.SpannerTestContext(t)
		ct := testclock.TestRecentTimeUTC

		mayMutate := func() error {
			return mutateInvocation(ctx, "inv", func(ctx context.Context, txn *spanner.ReadWriteTransaction) error {
				return nil
			})
		}

		Convey("no token", func() {
			err := mayMutate()
			So(err, ShouldErrLike, `missing "update-token" metadata value`)
			So(grpcutil.Code(err), ShouldEqual, codes.Unauthenticated)
		})

		Convey("with token", func() {
			const token = "update token"
			ctx = metadata.NewIncomingContext(ctx, metadata.Pairs(updateTokenMetadataKey, token))

			Convey(`no invocation`, func() {
				err := mayMutate()
				So(err, ShouldErrLike, `"invocations/inv" not found`)
				So(grpcutil.Code(err), ShouldEqual, codes.NotFound)
			})

			Convey(`with finalized invocation`, func() {
				testutil.MustApply(ctx, testutil.InsertInvocation("inv", pb.Invocation_COMPLETED, token, ct))
				err := mayMutate()
				So(err, ShouldErrLike, `"invocations/inv" is not active`)
				So(grpcutil.Code(err), ShouldEqual, codes.FailedPrecondition)
			})

			Convey(`with active invocation and different token`, func() {
				testutil.MustApply(ctx, testutil.InsertInvocation("inv", pb.Invocation_ACTIVE, "different token", ct))
				err := mayMutate()
				So(err, ShouldErrLike, `invalid update token`)
				So(grpcutil.Code(err), ShouldEqual, codes.PermissionDenied)
			})

			Convey(`with exceeded deadline`, func() {
				testutil.MustApply(ctx, testutil.InsertInvocation("inv", pb.Invocation_ACTIVE, token, ct))

				// Mock now to be after deadline.
				clock.Get(ctx).(testclock.TestClock).Add(2 * time.Hour)

				err := mayMutate()
				So(err, ShouldErrLike, `"invocations/inv" is not active`)
				So(grpcutil.Code(err), ShouldEqual, codes.FailedPrecondition)

				// Confirm the invocation has been updated.
				var state pb.Invocation_State
				var ft time.Time
				inv := span.InvocationID("inv")
				testutil.MustReadRow(ctx, "Invocations", inv.Key(), map[string]interface{}{
					"State":        &state,
					"FinalizeTime": &ft,
				})
				So(state, ShouldEqual, pb.Invocation_INTERRUPTED)
				So(ft, ShouldEqual, ct.Add(time.Hour))
			})

			Convey(`with active invocation and same token`, func() {
				testutil.MustApply(ctx, testutil.InsertInvocation("inv", pb.Invocation_ACTIVE, token, ct))

				err := mayMutate()
				So(err, ShouldBeNil)
			})
		})
	})
}

func TestReadInvocation(t *testing.T) {
	Convey(`ReadInvocationFull`, t, func() {
		ctx := testutil.SpannerTestContext(t)
		ct := testclock.TestRecentTimeUTC

		readInv := func() *pb.Invocation {
			txn := span.Client(ctx).ReadOnlyTransaction()
			defer txn.Close()

			inv, err := span.ReadInvocationFull(ctx, txn, "inv")
			So(err, ShouldBeNil)
			return inv
		}

		Convey(`completed`, func() {
			testutil.MustApply(ctx, testutil.InsertInvocation("inv", pb.Invocation_COMPLETED, "", ct))

			inv := readInv()
			expected := &pb.Invocation{
				Name:         "invocations/inv",
				State:        pb.Invocation_COMPLETED,
				CreateTime:   pbutil.MustTimestampProto(ct),
				Deadline:     pbutil.MustTimestampProto(ct.Add(time.Hour)),
				FinalizeTime: pbutil.MustTimestampProto(ct.Add(time.Hour)),
			}
			So(inv, ShouldResembleProto, expected)

			Convey(`with included invocations`, func() {
				testutil.MustApply(ctx,
					testutil.InsertInvocation("included0", pb.Invocation_COMPLETED, "", ct),
					testutil.InsertInvocation("included1", pb.Invocation_COMPLETED, "", ct),
					testutil.InsertInclusion("inv", "included0"),
					testutil.InsertInclusion("inv", "included1"),
				)

				inv := readInv()
				So(inv.IncludedInvocations, ShouldResemble, []string{"invocations/included0", "invocations/included1"})
			})
		})
	})
}
