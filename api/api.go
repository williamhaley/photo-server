package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/model"
)

// API handles all abstractions around the API.
type API struct {
	db     *datasource.Database
	schema graphql.Schema
}

// New returns a new instance of the API.
func New(db *datasource.Database) *API {
	return &API{
		db:     db,
		schema: newSchema(),
	}
}

func (api *API) BucketCounts() ([]interface{}, error) {
	log.Debug("[api:BucketCounts]")

	result := api.query(`{counts{year,month,totalCount}}`)
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			log.WithError(err)
		}
		return nil, errors.New("error retrieving bucket counts")
	}

	parsed := (result.Data.(map[string]interface{}))["counts"]

	return parsed.([]interface{}), nil
}

func (api *API) BucketPhotos(bucketID, after string) (interface{}, error) {
	log.Debugf("[api:BucketPhotos] %q", bucketID)

	result := api.query(fmt.Sprintf(`{
		yearMonthBucket(id:"%s"){
			photosConnection(first:20, after:"%s") {
				totalCount
				edges{
					node{
						uuid
						name
						date
					}
					cursor
				}
				pageInfo{
					endCursor
					hasNextPage
				}
			}
		}
	}`, bucketID, after))
	if len(result.Errors) > 0 {
		for _, err := range result.Errors {
			log.WithError(err)
		}
		return nil, fmt.Errorf("error retrieving photos for bucket %q", bucketID)
	}

	parsed := (result.Data.(map[string]interface{}))["yearMonthBucket"]

	return parsed, nil
}

// Query uses the provided query string to query GraphQL.
func (api *API) query(query string) *graphql.Result {
	return graphql.Do(graphql.Params{
		Schema:        api.schema,
		RequestString: query,
		Context:       context.WithValue(context.Background(), model.CtxDB, api.db),
	})
}
