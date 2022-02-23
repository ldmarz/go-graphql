package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/mercadolibre/fury_vis-sdk-go/pkg/items"
	"log"
	"net/http"
)

type postData struct {
	Query     string                 `json:"query"`
	Operation string                 `json:"operation"`
	Variables map[string]interface{} `json:"variables"`
}

func main() {
	// Schema

	statusType := graphql.NewEnum(graphql.EnumConfig{
		Name: "enum",
		Values: graphql.EnumValueConfigMap{
			"active": &graphql.EnumValueConfig{
				Value: "active",
			},
			"inactive": &graphql.EnumValueConfig{
				Value: "inactive",
			},
			"closed": &graphql.EnumValueConfig{
				Value: "closed",
			},
		},
	})

	attributesType := graphql.NewObject(graphql.ObjectConfig{
		Name: "attributes",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"value_name": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	itemType := graphql.NewObject(
		graphql.ObjectConfig{
			Name: "item",
			Fields: graphql.Fields{
				"id": &graphql.Field{
					Type: graphql.String,
				},
				"status": &graphql.Field{
					Type: statusType,
				},
				"attributes": &graphql.Field{
					Type: graphql.NewList(attributesType),
				},
			},
		},
	)

	rootField := graphql.Fields{
		"item": &graphql.Field{
			Type: itemType,
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				id, _ := p.Args["id"].(string)
				item, err := items.Get(context.TODO(), 123, id)
				if err != nil {
					return nil, err
				}

				return item, err
			},
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
		},
	}

	rootQuery := graphql.ObjectConfig{Name: "Query", Fields: rootField}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	// serve
	http.HandleFunc("/graphql", func(w http.ResponseWriter, req *http.Request) {
		var p postData
		if err := json.NewDecoder(req.Body).Decode(&p); err != nil {
			w.WriteHeader(400)
			return
		}
		result := graphql.Do(graphql.Params{
			Context:        req.Context(),
			Schema:         schema,
			RequestString:  p.Query,
			VariableValues: p.Variables,
			OperationName:  p.Operation,
		})

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			fmt.Printf("could not write result to response: %s", err)
		}
	})

	http.ListenAndServe(":8080", nil)
}
