package discovery

import (
	"context"
	"fmt"
	"log"

	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/pagination"
)

// ResourceExtractor is a function type that extracts resources from a pagination page
type ResourceExtractor func(pagination.Page) ([]interface{}, error)

// JobCreator is a function type that creates a Job from a resource
type JobCreator func(interface{}, string) (Job, error)

// DiscoverPaged is a generic helper function for discovering paged resources
// It handles common patterns like context cancellation, pagination, and error handling
func DiscoverPaged(
	ctx context.Context,
	client *gophercloud.ServiceClient,
	serviceName string,
	resourceType string,
	pager pagination.Pager,
	extract ResourceExtractor,
	createJob JobCreator,
) (<-chan Job, error) {
	jobChan := make(chan Job, 100)

	go func() {
		defer close(jobChan)

		err := pager.EachPage(func(page pagination.Page) (bool, error) {
			// Check context cancellation before processing page
			if err := ctx.Err(); err != nil {
				return false, err
			}

			// Extract resources from page
			resources, err := extract(page)
			if err != nil {
				return false, fmt.Errorf("failed to extract resources: %w", err)
			}

			// Process each resource
			for _, resource := range resources {
				// Check context cancellation before processing each resource
				if err := ctx.Err(); err != nil {
					return false, err
				}

				// Create job from resource
				job, err := createJob(resource, resourceType)
				if err != nil {
					log.Printf("Error creating job for %s/%s: %v", serviceName, resourceType, err)
					continue // Skip this resource but continue processing
				}

				// Send job to channel (with context check)
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				case jobChan <- job:
				}
			}

			return true, nil
		})

		if err != nil && err != context.Canceled {
			log.Printf("Error discovering %s/%s: %v", serviceName, resourceType, err)
		}
	}()

	return jobChan, nil
}

// SimpleJobCreator creates a helper function for simple job creation where
// the resource ID and project ID can be extracted using simple functions
func SimpleJobCreator(
	serviceName string,
	getID func(interface{}) string,
	getProjectID func(interface{}) string,
) JobCreator {
	return func(resource interface{}, resourceType string) (Job, error) {
		id := getID(resource)
		if id == "" {
			return Job{}, fmt.Errorf("resource ID is empty")
		}

		return Job{
			ResourceType: resourceType,
			ResourceID:   id,
			Resource:     resource,
			Service:      serviceName,
			ProjectID:    getProjectID(resource),
		}, nil
	}
}

