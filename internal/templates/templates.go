package templates

var ModelInit = `package {{.Entity | lower}}

import (
	"gitlab.silvertiger.tech/go-sdk/go-mongodb/collection"
	"{{.Module}}/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	{{.EntityLower}}Collection        *collection.MongoDBGenericCollection[{{.Entity}}]
	{{.EntityLower}}DeletedCollection *collection.MongoDBGenericCollection[{{.Entity}}]
	{{.EntityLower}}Repository        {{.Entity}}Repository
)

func Init(database *mongo.Database) error {
	{{.EntityLower}}Collection = collection.NewMongoDBGenericCollection[{{.Entity}}]("{{.DBName}}").(*collection.MongoDBGenericCollection[{{.Entity}}])
	{{.EntityLower}}Collection.SetDatabase(database)

	{{.EntityLower}}DeletedCollection = collection.NewMongoDBGenericCollection[{{.Entity}}]("{{.EntitySnake}}_deleted").(*collection.MongoDBGenericCollection[{{.Entity}}])
	{{.EntityLower}}DeletedCollection.SetDatabase(database)

	// Initialize repository
	{{.EntityLower}}Repository = &mongoRepository{}

	// Create indexes
	if err := createIndexes(); err != nil {
		return err
	}

	return nil
}

// GetRepository returns the initialized repository instance
func GetRepository() {{.Entity}}Repository {
	return {{.EntityLower}}Repository
}

func createIndexes() error {
	// TODO: Add your custom indexes here
	// Example:
	// err := {{.EntityLower}}Collection.CreateIndex(bson.D{
	//     {Key: "field_name", Value: 1},
	// }, &options.IndexOptions{
	//     Unique: utils.GetPointer(true),
	// })
	// if err != nil {
	//     return err
	// }

	return nil
}
`

var ModelRepository = `package {{.Entity | lower}}

// Repository defines the interface for {{.EntityLower}} operations
type Repository interface {
	Create(data *{{.Entity}}) (*{{.Entity}}, error)
	GetBy{{.Entity}}ID({{.EntityLower}}ID string) (*{{.Entity}}, error)
	List(filter interface{}, offset, limit int64, sort map[string]int) ([]*{{.Entity}}, error)
	Count(filter interface{}) (int64, error)
	UpdateBy{{.Entity}}ID({{.EntityLower}}ID string, data *{{.Entity}}) (*{{.Entity}}, error)
	DeleteBy{{.Entity}}ID({{.EntityLower}}ID string) error
}

// mongoRepository implements the Repository interface
type mongoRepository struct{}

func (r *mongoRepository) Create(data *{{.Entity}}) (*{{.Entity}}, error) {
	return {{.EntityLower}}Collection.InsertOne(data)
}

func (r *mongoRepository) GetBy{{.Entity}}ID({{.EntityLower}}ID string) (*{{.Entity}}, error) {
	return {{.EntityLower}}Collection.FindOne({{.Entity}}{ {{.Entity}}ID: {{.EntityLower}}ID})
}

func (r *mongoRepository) List(filter interface{}, offset, limit int64, sort map[string]int) ([]*{{.Entity}}, error) {
	return {{.EntityLower}}Collection.Find(filter, offset, limit, sort)
}

func (r *mongoRepository) Count(filter interface{}) (int64, error) {
	return {{.EntityLower}}Collection.Count(filter)
}

func (r *mongoRepository) UpdateBy{{.Entity}}ID({{.EntityLower}}ID string, data *{{.Entity}}) (*{{.Entity}}, error) {
	return {{.EntityLower}}Collection.UpdateOne({{.Entity}}{ {{.Entity}}ID: {{.EntityLower}}ID}, data)
}

// Delete implements Repository.Delete - Soft delete by moving to {{.EntitySnake}}_deleted collection
func (r *mongoRepository) DeleteBy{{.Entity}}ID({{.EntityLower}}ID string) error {
	{{.EntityLower}}, err := {{.EntityLower}}Collection.FindOne({{.Entity}}{ {{.Entity}}ID: {{.EntityLower}}ID})
	if err != nil {
		return err
	}
	_, err = {{.EntityLower}}DeletedCollection.InsertOne({{.EntityLower}})
	if err != nil {
		return err
	}
	return {{.EntityLower}}Collection.DeleteOne({{.Entity}}{ {{.Entity}}ID: {{.EntityLower}}ID})
}
`

var Action = `package action

import (
	"context"
	"{{.Module}}/{{.PkgPath}}"
)

type {{.Entity}}Service struct {
	Repo {{.Entity | lower}}.{{.Entity}}Repository
}

func (s *{{.Entity}}Service) Create(ctx context.Context, m *{{.Entity | lower}}.{{.Entity}}) error {
	return s.Repo.Create(ctx, m)
}

func (s *{{.Entity}}Service) Get(ctx context.Context, id string) (*{{.Entity | lower}}.{{.Entity}}, error) {
	return s.Repo.Get(ctx, id)
}

func (s *{{.Entity}}Service) Update(ctx context.Context, id string, m *{{.Entity | lower}}.{{.Entity}}) error {
	return s.Repo.Update(ctx, id, m)
}

func (s *{{.Entity}}Service) Delete(ctx context.Context, id string) error {
	return s.Repo.Delete(ctx, id)
}

func (s *{{.Entity}}Service) Query(ctx context.Context, q {{.Entity | lower}}.Query{{.Entity}}) ([]{{.Entity | lower}}.{{.Entity}}, int64, error) {
	return s.Repo.Query(ctx, q)
}
`

var API = `package api

import (
	"gitlab.silvertiger.tech/go-sdk/go-common/common"
	"gitlab.silvertiger.tech/go-sdk/go-common/request"
	"gitlab.silvertiger.tech/go-sdk/go-common/responder"
	"{{.Module}}/internal/action"
	"{{.Module}}/{{.PkgPath}}"
	constants "{{.Module}}/utils"
)

// Create{{.Entity}} creates a new {{.EntityLower}}
func Create{{.Entity}}(req request.APIRequest, res responder.APIResponder) error {
	var {{.EntityLower}}Data {{.Entity | lower}}.{{.Entity}}
	if err := req.ParseBody(&{{.EntityLower}}Data); err != nil {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "INVALID_REQUEST_BODY", "Failed to parse request body: "+err.Error()))
	}

	// TODO: Add validation for required fields
	// Example:
	// if {{.EntityLower}}Data.{{.Entity}}ID == "" {
	//     return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "VALIDATION_FAILED", "{{.EntityLower}}_id is required"))
	// }

	response := action.Create{{.Entity}}(&{{.EntityLower}}Data)
	return res.Respond(response)
}

// Get{{.Entity}}By{{.Entity}}ID retrieves a {{.EntityLower}} by its {{.Entity}}ID
func Get{{.Entity}}By{{.Entity}}ID(req request.APIRequest, res responder.APIResponder) error {
	{{.EntityLower}}ID := req.GetParam(constants.Param{{.Entity}}ID)
	if {{.EntityLower}}ID == "" {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "VALIDATION_FAILED", "{{.EntityLower}}_id parameter is required"))
	}

	response := action.Get{{.Entity}}By{{.Entity}}ID({{.EntityLower}}ID)
	return res.Respond(response)
}

// Query{{.EntityPlural}} retrieves a list of {{.EntityLower}}s with optional filtering
func Query{{.EntityPlural}}(req request.APIRequest, res responder.APIResponder) error {
	var query common.Query[{{.Entity | lower}}.{{.Entity}}]
	if err := req.ParseBody(&query); err != nil {
		return res.Respond(common.FromError(err))
	}

	return res.Respond(action.List{{.EntityPlural}}(&query))
}

// Update{{.Entity}} updates an existing {{.EntityLower}}
func Update{{.Entity}}(req request.APIRequest, res responder.APIResponder) error {
	{{.EntityLower}}ID := req.GetParam(constants.Param{{.Entity}}ID)
	if {{.EntityLower}}ID == "" {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "VALIDATION_FAILED", "id parameter is required"))
	}

	var {{.EntityLower}}Data {{.Entity | lower}}.{{.Entity}}
	if err := req.ParseBody(&{{.EntityLower}}Data); err != nil {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "INVALID_REQUEST_BODY", "Failed to parse request body: "+err.Error()))
	}

	response := action.Update{{.Entity}}({{.EntityLower}}ID, &{{.EntityLower}}Data)
	return res.Respond(response)
}

// Delete{{.Entity}} deletes a {{.EntityLower}} by ID
func Delete{{.Entity}}(req request.APIRequest, res responder.APIResponder) error {
	{{.EntityLower}}ID := req.GetParam(constants.Param{{.Entity}}ID)
	if {{.EntityLower}}ID == "" {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "VALIDATION_FAILED", "id parameter is required"))
	}

	response := action.Delete{{.Entity}}({{.EntityLower}}ID)
	return res.Respond(response)
}
`

var Client = `package client

import (
	"gitlab.silvertiger.tech/go-sdk/go-common/common"
	"{{.Module}}/{{.PkgPath}}"
)

// Create{{.Entity}} creates a new {{.EntityLower}}
func (c *BackendServiceClient) Create{{.Entity}}(data *{{.Entity | lower}}.{{.Entity}}) *common.APIResponse[*{{.Entity | lower}}.{{.Entity}}] {
	response := &common.APIResponse[*{{.Entity | lower}}.{{.Entity}}]{}
	c.makeRequest("POST", "/v1/{{.EntityLower}}", nil, data, response)

	return response
}

// Get{{.Entity}}By{{.Entity}}ID retrieves a {{.EntityLower}} by its {{.EntityLower}}_id
func (c *BackendServiceClient) Get{{.Entity}}(id string) *common.APIResponse[*{{.Entity | lower}}.{{.Entity}}] {
	params := map[string]string{
		"{{.EntityLower}}_id": id,
	}
	response := &common.APIResponse[*{{.Entity | lower}}.{{.Entity}}]{}
	c.makeRequest("GET", "/v1/{{.EntityLower}}", params, nil, response)

	return response
}

// List{{.EntityPlural}} retrieves a list of {{.EntityLower}}s with filtering
func (c *BackendServiceClient) List{{.EntityPlural}}(query *common.Query[{{.Entity | lower}}.{{.Entity}}]) *common.APIResponse[*{{.Entity | lower}}.{{.Entity}}] {
	response := &common.APIResponse[*{{.Entity | lower}}.{{.Entity}}]{}
	c.makeRequest("QUERY", "/v1/{{.EntityLower}}s", nil, query, response)

	return response
}

// Update{{.Entity}} updates an existing {{.EntityLower}}
func (c *BackendServiceClient) Update{{.Entity}}(id string, data *{{.Entity | lower}}.{{.Entity}}) *common.APIResponse[*{{.Entity | lower}}.{{.Entity}}] {
	params := map[string]string{
		"{{.EntityLower}}_id": id,
	}
	response := &common.APIResponse[*{{.Entity | lower}}.{{.Entity}}]{}
	c.makeRequest("PUT", "/v1/{{.EntityLower}}", params, data, response)

	return response
}

// Delete{{.Entity}} deletes a {{.EntityLower}} by ID
func (c *BackendServiceClient) Delete{{.Entity}}(id string) *common.APIResponse[any] {
	params := map[string]string{
		"{{.EntityLower}}_id": id,
	}
	response := &common.APIResponse[any]{}
	c.makeRequest("DELETE", "/v1/{{.EntityLower}}", params, nil, response)

	return response
}
`

var MainRouter = `
	// Register {{.Entity}} API routes
	// Core CRUD operations
	server.SetHandler(common.APIMethod.POST, "/v1/{{.EntityLower}}", api.Create{{.Entity}})
	server.SetHandler(common.APIMethod.GET, "/v1/{{.EntityLower}}", api.Get{{.Entity}}By{{.Entity}}ID)
	server.SetHandler(common.APIMethod.QUERY, "/v1/{{.EntityLower}}s", api.Query{{.EntityPlural}})
	server.SetHandler(common.APIMethod.PUT, "/v1/{{.EntityLower}}", api.Update{{.Entity}})
	server.SetHandler(common.APIMethod.DELETE, "/v1/{{.EntityLower}}", api.Delete{{.Entity}})
`

var MainInit = `	{{.Entity | lower}}.Init(database)`

var MainGo = `package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"

	"gitlab.silvertiger.tech/go-sdk/go-backend/backend"
	"gitlab.silvertiger.tech/go-sdk/go-cache/rcache"
	"gitlab.silvertiger.tech/go-sdk/go-common/common"
	"gitlab.silvertiger.tech/go-sdk/go-common/request"
	"gitlab.silvertiger.tech/go-sdk/go-common/responder"
	"gitlab.silvertiger.tech/go-sdk/go-common/server"
	"gitlab.silvertiger.tech/go-sdk/go-mongodb/client"
	"{{.Module}}/internal/api"
	"{{.Module}}/internal/conf"
	// Import all model packages for initialization
	{{range .Entities}}"{{$.Module}}/{{.PkgPath}}"
	{{end}}
)

type infoData struct {
	Service     string    ` + "`json:\"service\"`" + `
	Environment string    ` + "`json:\"environment\"`" + `
	Version     string    ` + "`json:\"version\"`" + `
	StartTime   time.Time ` + "`json:\"startTime\"`" + `
}

var globalInfo *infoData

func info(req request.APIRequest, res responder.APIResponder) error {
	return res.Respond(&common.APIResponse[infoData]{
		Status:  common.APIStatus.Ok,
		Data:    []infoData{*globalInfo},
		Message: "Service runs normally.",
	})
}

func onMainDBConnected(database *mongo.Database) error {
	fmt.Println("Successfully connected to MongoDB!")

	// Initialize all models
{{range .Entities}}	{{.Name | lower}}.Init(database)
{{end}}
	return nil
}

func initMongoClient() {
	// Create a configuration
	mainDBConfig := client.Configuration{
		Username:      conf.Conf.DBConfig.MainDB.Username,
		Password:      conf.Conf.DBConfig.MainDB.Password,
		Address:       conf.Conf.DBConfig.MainDB.Address,
		DBName:        conf.Conf.DBConfig.MainDB.DBName,
		AuthDB:        conf.Conf.DBConfig.MainDB.AuthDB,
		AuthMechanism: conf.Conf.DBConfig.MainDB.AuthMechanism,
		DoWriteTest:   true,
	}
	// Create a MongoDB client
	mongoClient := client.NewMongoClient(conf.Conf.BackendConfig.BackendApp.ServiceName, mainDBConfig, onMainDBConnected)

	// Connect to the database
	err := mongoClient.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
}

func initRedisClient() {
	if conf.Conf.DBConfig.KeyDB == nil {
		return
	}

	redisConfig := &rcache.Configuration{
		Name: conf.Conf.BackendConfig.BackendApp.ServiceName,
		Option: &redis.Options{
			// TODO: Use proper config values
			Addr:     conf.Conf.DBConfig.KeyDB.Address,
			Username: conf.Conf.DBConfig.KeyDB.Username,
			Password: conf.Conf.DBConfig.KeyDB.Password,
			DB:       conf.Conf.DBConfig.KeyDB.DB,
		},
		Required: true, // Panic if Redis connection fails
	}

	// Create Redis client
	_, err := rcache.NewClient(redisConfig)
	if err != nil {
		// Log error but don't fail - fallback to no cache
		fmt.Printf("Failed to initialize Redis client for cache: %v\n", err)
		return
	}
}

func main() {
	// load config from env
	// don't use func init default because this project will be used as a library
	conf.NewConfig()

	globalInfo = &infoData{
		Service:     conf.Conf.BackendConfig.BackendApp.ServiceName,
		Version:     conf.Conf.Version,
		Environment: conf.Conf.Env,
		StartTime:   time.Now(),
	}

	initMongoClient()
	initRedisClient()

	// Create a new server
	server := server.NewServer(server.ServerConfig{
		Protocol: conf.Conf.Protocol,
	})

	// Initialize the backend
	backend.NewBackend(server, conf.Conf.BackendConfig.BackendApp.Username, conf.Conf.BackendConfig.BackendApp.Password, conf.Conf.BackendConfig.BackendApp.SecretKey)

	server.SetHandler(common.APIMethod.GET, "/api-info", info)

	// Register API routes for all entities
{{range .Entities}}
	// Register {{.Name}} API routes
	// Core CRUD operations
	server.SetHandler(common.APIMethod.POST, "/v1/{{.Name | lower}}", api.Create{{.Name}})
	server.SetHandler(common.APIMethod.GET, "/v1/{{.Name | lower}}", api.Get{{.Name}}By{{.Name}}ID)
	server.SetHandler(common.APIMethod.QUERY, "/v1/{{.Name | lower}}s", api.Query{{.Plural}})
	server.SetHandler(common.APIMethod.PUT, "/v1/{{.Name | lower}}", api.Update{{.Name}})
	server.SetHandler(common.APIMethod.DELETE, "/v1/{{.Name | lower}}", api.Delete{{.Name}})
{{end}}
	// Expose the server
	server.Expose(conf.Conf.Port)

	// Use a WaitGroup to keep the main function from exiting
	var wg sync.WaitGroup
	wg.Add(1)
	go server.Start(&wg)

	wg.Wait()
}
`
