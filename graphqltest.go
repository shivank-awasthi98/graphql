package main

import (
	"context"
	"fmt"

	sa "graphQlTest/sa"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/graphql-go/graphql"
	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
	"google.golang.org/genproto/protobuf/field_mask"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "12345"
	dbname   = "suser"
)

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

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)
	fmt.Println(psqlInfo)
	db, err := gorm.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("errrrror", err)
		return
	}
	defer db.Close()
	db.CreateTable(&sa.UserORM{})
	fmt.Println("connected")

	r := mux.NewRouter()

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
						users, _ := sa.DefaultListUser(context.Background(), db)

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
					users, _ := sa.DefaultListUser(context.Background(), db)
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
			"createuser": &graphql.Field{
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
					var user sa.User

					user.Username = params.Args["username"].(string)
					user.Email = params.Args["email"].(string)
					user.Password = params.Args["password"].(string)
					_, err := sa.DefaultCreateUser(context.Background(), &user, db)
					if err != nil {
						panic(err)
					}
					fmt.Println(user)
					return user, nil
				},
			},

			"updateuser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"username": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {

					id, ok := params.Args["id"].(int)
					if ok {

						users, _ := sa.DefaultListUser(context.Background(), db)

						for _, user := range users {
							if int(user.Id) == id {
								fmt.Println(user)
								user.Username = params.Args["username"].(string)
								user, err = sa.DefaultPatchUser(context.Background(), user, &field_mask.FieldMask{
									Paths: []string{"Username"},
								}, db)

								return user, nil
							}
						}
					}
					return nil, nil

				},
			},

			"deleteuser": &graphql.Field{
				Type: userType,
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					users, err := sa.DefaultListUser(context.Background(), db)
					if err != nil {
						panic(err)
					}
					for _, user := range users {
						if int(user.Id) == params.Args["id"] {
							fmt.Println(user)
							err := sa.DefaultDeleteUser(context.Background(), user, db)
							if err != nil {
								panic(err)
							}
							return user, nil
						}

					}
					return nil, err
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

	r.HandleFunc("/graphql", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)
		result := executeQuery(r.URL.Query().Get("query"), schema)
		fmt.Println(result)
	})
	fmt.Println(sa.DefaultListUser(context.Background(), db))
	http.ListenAndServe(":80", r)
}
