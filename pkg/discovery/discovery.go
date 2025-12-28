package discovery

import (
	"log"

	"github.com/OpenStack-Policy-Agent/OSPA/pkg/engine"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
)

// DiscoverServers lists all servers and pushes them to the jobs channel.
func DiscoverServers(client *gophercloud.ServiceClient, allTenants bool, jobs chan<- engine.Job) {
	opts := servers.ListOpts{
		AllTenants: allTenants, // Scan the whole cloud (requires admin)
	}

	log.Println("Starting Server Discovery...")

	pager := servers.List(client, opts)
	err := pager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}

		for _, s := range serverList {
			jobs <- engine.Job{Server: s}
		}
		return true, nil
	})

	if err != nil {
		log.Printf("Error listing servers: %v", err)
	}

	log.Println("Discovery complete. Closing job channel.")
	close(jobs) // Signal workers that no more jobs are coming
}

 