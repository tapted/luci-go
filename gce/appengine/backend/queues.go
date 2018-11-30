// Copyright 2018 The LUCI Authors.
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

package backend

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"

	"go.chromium.org/gae/service/datastore"
	"go.chromium.org/luci/appengine/tq"
	"go.chromium.org/luci/common/data/rand/mathrand"
	"go.chromium.org/luci/common/errors"
	"go.chromium.org/luci/common/logging"

	"go.chromium.org/luci/gce/api/config/v1"
	"go.chromium.org/luci/gce/api/tasks/v1"
	"go.chromium.org/luci/gce/appengine/model"
)

// getSuffix returns a random suffix to use when naming a GCE instance.
func getSuffix(c context.Context) string {
	const allowed = "abcdefghijklmnopqrstuvwxyz0123456789"
	suf := make([]byte, 4)
	for i := range suf {
		suf[i] = allowed[mathrand.Intn(c, len(allowed))]
	}
	return string(suf)
}

// createQueue is the name of the create task handler queue.
const createQueue = "create-instance"

// create creates a GCE instance.
func create(c context.Context, payload proto.Message) error {
	task, ok := payload.(*tasks.Create)
	switch {
	case !ok:
		return errors.Reason("unexpected payload %q", payload).Err()
	case task.GetId() == "":
		return errors.Reason("ID is required").Err()
	}
	vm := &model.VM{
		ID: task.Id,
	}
	if err := datastore.Get(c, vm); err != nil {
		return errors.Annotate(err, "failed to fetch VM").Err()
	}
	if vm.Hostname == "" {
		// Generate a new hostname and record it so future calls are idempotent.
		hostname := fmt.Sprintf("%s-%s", vm.ID, getSuffix(c))
		if err := datastore.RunInTransaction(c, func(c context.Context) error {
			if err := datastore.Get(c, vm); err != nil {
				return errors.Annotate(err, "failed to fetch VM").Err()
			}
			// Double-check inside transaction. Hostname may already be generated.
			if vm.Hostname == "" {
				vm.Hostname = hostname
				if err := datastore.Put(c, vm); err != nil {
					return errors.Annotate(err, "failed to store VM").Err()
				}
			}
			return nil
		}, nil); err != nil {
			return err
		}
	}
	if vm.URL != "" {
		logging.Debugf(c, "VM exists: %q", vm.URL)
		return nil
	}
	// Generate a request ID based on the hostname.
	// Ensures duplicate operations aren't created in GCE.
	rID := uuid.NewSHA1(uuid.Nil, []byte(vm.Hostname))
	srv := compute.NewInstancesService(getCompute(c))
	call := srv.Insert(vm.Attributes.GetProject(), vm.Attributes.GetZone(), vm.GetInstance())
	op, err := call.RequestId(rID.String()).Context(c).Do()
	if err != nil {
		for _, err := range err.(*googleapi.Error).Errors {
			logging.Errorf(c, "%s", err.Message)
		}
		return errors.Reason("failed to create instance").Err()
	}
	logging.Infof(c, "operation %q", op)
	// TODO(smut): Check operation status.
	return nil
}

// ensureQueue is the name of the ensure task handler queue.
const ensureQueue = "ensure-vm"

// ensure creates or updates a given VM.
func ensure(c context.Context, payload proto.Message) error {
	task, ok := payload.(*tasks.Ensure)
	switch {
	case !ok:
		return errors.Reason("unexpected payload %q", payload).Err()
	case task.GetId() == "":
		return errors.Reason("ID is required").Err()
	}
	return datastore.RunInTransaction(c, func(c context.Context) error {
		vm := &model.VM{
			ID: task.Id,
		}
		if err := datastore.Get(c, vm); err != nil && err != datastore.ErrNoSuchEntity {
			return errors.Annotate(err, "failed to fetch VM").Err()
		}
		if task.Attributes != nil {
			vm.Attributes = *task.Attributes
		}
		vm.Prefix = task.Prefix
		if err := datastore.Put(c, vm); err != nil {
			return errors.Annotate(err, "failed to store VM").Err()
		}
		return nil
	}, nil)
}

// expandQueue is the name of the expand task handler queue.
const expandQueue = "expand-config"

// expand creates task queue tasks to process each VM in the given VMs block.
func expand(c context.Context, payload proto.Message) error {
	task, ok := payload.(*tasks.Expand)
	switch {
	case !ok:
		return errors.Reason("unexpected payload %q", payload).Err()
	case task.GetId() == "":
		return errors.Reason("ID is required").Err()
	}
	vms, err := getConfig(c).GetVMs(c, &config.GetVMsRequest{Id: task.Id})
	if err != nil {
		return errors.Annotate(err, "failed to get VMs block").Err()
	}
	logging.Debugf(c, "found %d VMs", vms.Amount)
	t := make([]*tq.Task, vms.Amount)
	for i := int32(0); i < vms.Amount; i++ {
		t[i] = &tq.Task{
			Payload: &tasks.Ensure{
				Id:         fmt.Sprintf("%s-%d", task.Id, i),
				Attributes: vms.Attributes,
			},
		}
	}
	if err := getDispatcher(c).AddTask(c, t...); err != nil {
		return errors.Annotate(err, "failed to schedule tasks").Err()
	}
	return nil
}