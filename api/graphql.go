package api

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/graph-gophers/dataloader"
	"github.com/graphql-go/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/model"
)

type Result struct {
	Records    interface{}
	LastCursor string
	TotalCount int
	HasMore    bool
}

func NewResult(records interface{}, lastCursor string, count int, hasMore bool) *Result {
	return &Result{
		Records:    records,
		LastCursor: lastCursor,
		TotalCount: count,
		HasMore:    hasMore,
	}
}

// https://github.com/graphql-go/graphql-dataloader-example/blob/master/main.go
var handleError = func(err error) []*dataloader.Result {
	var results []*dataloader.Result
	var result dataloader.Result
	result.Error = err
	results = append(results, &result)
	return results
}

var yearMonthBucketLoader = dataloader.NewBatchedLoader(
	func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
		db := ctx.Value(model.CtxDB).(*datasource.Database)

		yearMonthDates := keys.Keys()
		results := make([]*dataloader.Result, len(yearMonthDates))

		log.Debugf("[graphql:yearMonthBucketLoader]: %q", yearMonthDates)

		buckets, err := db.DateBucketsForIds(yearMonthDates...)
		if err != nil {
			return handleError(err)
		}

		for index, bucket := range buckets {
			results[index] = &dataloader.Result{
				Data:  bucket,
				Error: nil,
			}
		}

		return results
	},
)

var photosConnectionType = newConnectionResult(photoType)

var photoType = graphql.NewObject(graphql.ObjectConfig{
	Name: "photo",
	Fields: graphql.Fields{
		"uuid": &graphql.Field{
			Type: graphql.String,
		},
		"path": &graphql.Field{
			Type: graphql.String,
		},
		"name": &graphql.Field{
			Type: graphql.String,
		},
		"date": &graphql.Field{
			Type: graphql.DateTime,
		},
		"cursor": &graphql.Field{
			Type: graphql.String,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				photo := params.Source.(*model.Photo)
				return photo.Cursor(), nil
			},
		},
	},
})

var yearMonthBucketType = graphql.NewObject(graphql.ObjectConfig{
	Name: "yearMonthBucket",
	Fields: graphql.Fields{
		"year": &graphql.Field{
			Type: graphql.Int,
		},
		"month": &graphql.Field{
			Type: graphql.Int,
		},
		"totalCount": &graphql.Field{
			Type: graphql.Int,
		},
		"photosConnection": &graphql.Field{
			Type: photosConnectionType,
			Args: graphql.FieldConfigArgument{
				"first": &graphql.ArgumentConfig{
					Type:         graphql.Int,
					DefaultValue: 10,
				},
				"after": &graphql.ArgumentConfig{
					Type:         graphql.String,
					DefaultValue: "",
				},
			},
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				yearMonthBucket := params.Source.(*model.YearMonthBucket)
				limit := params.Args["first"].(int)
				decodedCursor, err := base64.StdEncoding.DecodeString(params.Args["after"].(string))
				if err != nil {
					log.WithError(err).Errorf("error decoding cursor %q", params.Args["after"])
					return nil, err
				}
				after := string(decodedCursor)

				year := yearMonthBucket.Year
				month := yearMonthBucket.Month

				log.Debugf("[graphql:resolvePhotosForYearMonth]: %d-%d %q", year, month, after)

				db := params.Context.Value(model.CtxDB).(*datasource.Database)

				photos, hasMore, err := db.AllPhotos(year, month, limit, after)
				if err != nil {
					log.WithError(err)
					return nil, err
				}

				cursor := ""
				if len(photos) > 0 {
					lastPhoto := photos[len(photos)-1]
					cursor = lastPhoto.Cursor()
				}

				count, err := db.PhotosCount(year, month)
				if err != nil {
					log.WithError(err)
					return nil, err
				}

				return NewResult(photos, cursor, count, hasMore), nil
			},
		},
	},
})

func newConnectionResult(nodeType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: fmt.Sprintf("%sConnectionResult", nodeType.Name()),
		Fields: graphql.Fields{
			"totalCount": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					result := params.Source.(*Result)
					return result.TotalCount, nil
				},
			},
			"edges": &graphql.Field{
				Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
					Name: "edge",
					Fields: graphql.Fields{
						"node": &graphql.Field{
							Type: nodeType,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								return params.Source, nil
							},
						},
						"cursor": &graphql.Field{
							Type: graphql.String,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								switch record := params.Source.(type) {
								case model.Cursorable:
									return record.Cursor(), nil
								}
								return "", nil
							},
						},
					},
				})),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					result := params.Source.(*Result)
					return result.Records, nil
				},
			},

			"pageInfo": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "pageInfo",
					Fields: graphql.Fields{
						"endCursor": &graphql.Field{
							Type: graphql.String,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								result := params.Source.(*Result)
								return result.LastCursor, nil
							},
						},
						"hasNextPage": &graphql.Field{
							Type: graphql.Boolean,
							Resolve: func(params graphql.ResolveParams) (interface{}, error) {
								result := params.Source.(*Result)
								return result.HasMore, nil
							},
						},
					},
				}),
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					return params.Source, nil
				},
			},
		},
	})
}

func newSchema() graphql.Schema {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query: graphql.NewObject(
			graphql.ObjectConfig{
				Name: "RootQuery",
				Fields: graphql.Fields{
					"counts": &graphql.Field{
						Type: graphql.NewList(yearMonthBucketType),
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							db := params.Context.Value(model.CtxDB).(*datasource.Database)

							return db.SkeletonMetaData()
						},
					},

					"yearMonthBucket": &graphql.Field{
						Type: yearMonthBucketType,
						Args: graphql.FieldConfigArgument{
							"id": &graphql.ArgumentConfig{
								Type:         graphql.String,
								DefaultValue: 0,
							},
						},
						Resolve: func(params graphql.ResolveParams) (interface{}, error) {
							id := params.Args["id"].(string)
							keys := dataloader.NewKeysFromStrings([]string{id})
							thunk := yearMonthBucketLoader.Load(params.Context, keys[0])
							return func() (interface{}, error) {
								return thunk()
							}, nil
						},
					},
				},
			},
		),
	})
	if err != nil {
		log.WithError(err).Fatal("failed to create new schema")
	}

	return schema
}
