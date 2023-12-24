package main

import (
	"context"
	"entgo.io/ent/dialect/sql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp/v3"
	"github.com/rs/cors"
	"graphql-test-api/ent"
	"graphql-test-api/ent/migrate"
	"graphql-test-api/graph"
	"graphql-test-api/graph/generated"
	"log"
	"net/http"
	"os"
	"time"
)

const defaultPort = "8080"

func Open() (*ent.Client, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	pp.Print(os.Getenv("DSN"))
	drv, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		return nil, err
	}
	// Get the underlying sql.DB object of the driver.
	db := drv.DB()
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Hour)
	var entOptions []ent.Option
	entOptions = append(entOptions, ent.Debug())
	entOptions = append(entOptions, ent.Driver(drv))
	return ent.NewClient(entOptions...), nil
}

func main() {

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	client, err := Open()
	if err != nil {
		log.Fatalf("failed opening connection to sqlite: %v", err)
	}
	defer client.Close()
	ctx := context.Background()
	// マイグレーションの実行
	err = client.Schema.Create(
		ctx,
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	)
	if err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{
		Client: client,
	}}))
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowCredentials: true,
	})
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", c.Handler(srv))

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
