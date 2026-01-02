package discovery

import (
	"context"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/volumes"
	"github.com/gophercloud/gophercloud/openstack/blockstorage/v3/snapshots"
	"github.com/gophercloud/gophercloud/pagination"
)

// BlockStorageVolumeDiscoverer discovers Cinder volumes
type BlockStorageVolumeDiscoverer struct{}

func (d *BlockStorageVolumeDiscoverer) ResourceType() string {
	return "volume"
}

func (d *BlockStorageVolumeDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
	opts := volumes.ListOpts{}
	if allTenants {
		opts.AllTenants = true
	}
	pager := volumes.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		volumeList, err := volumes.ExtractVolumes(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(volumeList))
		for i := range volumeList {
			resources[i] = volumeList[i]
		}
		return resources, nil
	}

	createJob := SimpleJobCreator(
		"cinder",
		func(r interface{}) string {
			return r.(volumes.Volume).ID
		},
		func(r interface{}) string {
			return r.(volumes.Volume).OsVolTenantAttr.TenantID
		},
	)

	return DiscoverPaged(ctx, client, "cinder", d.ResourceType(), pager, extract, createJob)
}

// BlockStorageSnapshotDiscoverer discovers Cinder snapshots
type BlockStorageSnapshotDiscoverer struct{}

func (d *BlockStorageSnapshotDiscoverer) ResourceType() string {
	return "snapshot"
}

func (d *BlockStorageSnapshotDiscoverer) Discover(ctx context.Context, client *gophercloud.ServiceClient, allTenants bool) (<-chan Job, error) {
	opts := snapshots.ListOpts{}
	if allTenants {
		opts.AllTenants = true
	}
	pager := snapshots.List(client, opts)

	extract := func(page pagination.Page) ([]interface{}, error) {
		snapshotList, err := snapshots.ExtractSnapshots(page)
		if err != nil {
			return nil, err
		}
		resources := make([]interface{}, len(snapshotList))
		for i := range snapshotList {
			resources[i] = snapshotList[i]
		}
		return resources, nil
	}

	createJob := SimpleJobCreator(
		"cinder",
		func(r interface{}) string {
			return r.(snapshots.Snapshot).ID
		},
		func(r interface{}) string {
			return r.(snapshots.Snapshot).OsExtendedSnapshotAttributes.ProjectID
		},
	)

	return DiscoverPaged(ctx, client, "cinder", d.ResourceType(), pager, extract, createJob)
}

