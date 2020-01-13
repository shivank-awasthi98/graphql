package main

import (
	"fmt"
	"github.com/graphql-go/graphql"

	"net/http"
)

type User struct {
	Id       int    `json:"id" `
	UserName string `json:"username" `
	Email    string `json:"email" `
	Password string `json:"password" `
}

func executeQuery(query string, schema graphql.Schema) *graphql.Result {
	result := graphql.Do(graphql.Params{
		Schema:        schema,
		RequestString: query,
	})
	if len(result.Errors) > 0 {
		fmt.Printf("errors: %v", result.Errors)
	}
	return result
}

func main() {
	var users []User = []User{
		User{
			Id:       1,
			UserName: "sss",
			Email:    "sss@email",
			Password: "sss@password",
		},
	}

	userType := graphql.NewObject(graphql.ObjectConfig{
		Name: "user",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.Int,
			},
			"username": &graphql.Field{
				Type: graphql.String,
			},
			"email": &graphql.Field{
				Type: graphql.String,
			},
			"password": &graphql.Field{
				Type: graphql.String,
			},
		},
	})
	RootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"user": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					id, ok := p.Args["id"].(int)
					if ok {

						for _, user := range users {
							if int(user.Id) == id {
								fmt.Println(user)
								return user, nil
							}
						}
					}
					return nil, nil
				},
			},
			"users": &graphql.Field{
				Type: graphql.NewList(userType),

				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					fmt.Println(users)
					return users, nil
				},
			},
		},
	},
	)

	Mutations := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"Createuser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"username": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"email": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"password": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					var user User
					user.Id = params.Args["id"].(int)
					user.UserName = params.Args["username"].(string)
					user.Email = params.Args["email"].(string)
					user.Password = params.Args["password"].(string)
					users = append(users, user)
					fmt.Println(users)
					return users, nil
				},
			},
		},
	},
	)

	schema, _ := graphql.NewSchema(graphql.SchemaConfig{
		Query:    RootQuery,
		Mutation: Mutations,
	})
	fmt.Println(schema)

	http.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		result := executeQuery(r.URL.Query().Get("query"), schema)
		fmt.Println(result)
	})
	http.ListenAndServe(":80", nil)
}
