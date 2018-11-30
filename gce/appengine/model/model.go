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

package model

import (
	"google.golang.org/api/compute/v1"

	"go.chromium.org/gae/service/datastore"

	"go.chromium.org/luci/gce/api/config/v1"
)

// VMsKind is a VMs entity's kind in the datastore.
const VMsKind = "VMs"

// VMs is a root entity representing a configured block of VMs.
// VM entities should be created for each VMs entity.
type VMs struct {
	// _extra is where unknown properties are put into memory.
	// Extra properties are not written to the datastore.
	_extra datastore.PropertyMap `gae:"-,extra"`
	// _kind is the entity's kind in the datastore.
	_kind string `gae:"$kind,VMs"`
	// ID is the unique identifier for this VMs block.
	ID string `gae:"$id"`
	// Config is the config.Block representation of this entity.
	Config config.Block `gae:"config"`
}

// VMKind is a VM entity's kind in the datastore.
const VMKind = "VM"

// VM is a root entity representing a configured VM.
// GCE instances should be created for each VM entity.
type VM struct {
	// _extra is where unknown properties are put into memory.
	// Extra properties are not written to the datastore.
	_extra datastore.PropertyMap `gae:"-,extra"`
	// _kind is the entity's kind in the datastore.
	_kind string `gae:"$kind,VM"`
	// ID is the unique identifier for this VM.
	ID string `gae:"$id"`
	// Attributes is the config.VM describing the GCE instance to create.
	Attributes config.VM `gae:"attributes"`
	// Hostname is the short hostname of the GCE instance to create.
	Hostname string `gae:"hostname"`
	// Prefix is the prefix to use when naming the GCE instance.
	Prefix string `gae:"prefix"`
	// URL is the URL of the created GCE instance.
	URL string `gae:"url"`
}

// GetInstance returns a *compute.Instance representation of this VM.
func (vm *VM) GetInstance() *compute.Instance {
	inst := &compute.Instance{
		Name:        vm.Hostname,
		MachineType: vm.Attributes.GetMachineType(),
		// One network interface is required, but GCE can infer all defaults.
		// TODO(smut): Allow the config to optionally configure NICs.
		NetworkInterfaces: []*compute.NetworkInterface{
			{},
		},
	}
	inst.Disks = make([]*compute.AttachedDisk, len(vm.Attributes.GetDisk()))
	for i, disk := range vm.Attributes.GetDisk() {
		inst.Disks[i] = &compute.AttachedDisk{
			// AutoDelete deletes the disk when the instance is deleted.
			AutoDelete: true,
			InitializeParams: &compute.AttachedDiskInitializeParams{
				DiskSizeGb:  disk.Size,
				DiskType:    disk.Type,
				SourceImage: disk.Image,
			},
		}
	}
	if len(inst.Disks) > 0 {
		// GCE requires the first disk to be the boot disk.
		inst.Disks[0].Boot = true
	}
	return inst
}