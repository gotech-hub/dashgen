package templates

var ModelInit = `package {{.Entity | lower}}

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"gitlab.silvertiger.tech/go-sdk/go-mongodb/collection"
	"{{.Module}}/internal/utils"
)

var (
	{{.EntityLower}}Collection        *collection.MongoDBGenericCollection[{{.Entity}}]
	{{.EntityLower}}DeletedCollection *collection.MongoDBGenericCollection[{{.Entity}}]
	{{.EntityLower}}Repository        Repository
)

func Init(database *mongo.Database) error {
	{{.EntityLower}}DeletedCollection = collection.NewMongoDBGenericCollection[{{.Entity}}]("{{.EntitySnake}}_deleted").(*collection.MongoDBGenericCollection[{{.Entity}}])
	{{.EntityLower}}DeletedCollection.SetDatabase(database)

	{{.EntityLower}}Collection = collection.NewMongoDBGenericCollection[{{.Entity}}]("{{.DBName}}").(*collection.MongoDBGenericCollection[{{.Entity}}])
	{{.EntityLower}}Collection.SetDatabase(database)

	// Initialize repository
	{{.EntityLower}}Repository = &mongoRepository{}

	// Create indexes
	if err := createIndexes(); err != nil {
		return err
	}

	return nil
}

// GetRepository returns the initialized repository instance
func GetRepository() Repository {
	return {{.EntityLower}}Repository
}

func createIndexes() error {
{{if hasIndexes .Fields .Indexes}}{{generateIndexes .Fields .Indexes .EntityLower}}{{else}}	// No indexes defined{{end}}

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
	"gitlab.silvertiger.tech/go-sdk/go-common/common"
	"{{.Module}}/{{.PkgPath}}"
)

// Create{{.Entity}} creates a new {{.EntityLower}}
func Create{{.Entity}}(data *{{.EntityLower}}.{{.Entity}}) *common.APIResponse[*{{.EntityLower}}.{{.Entity}}] {
	repo := {{.EntityLower}}.GetRepository()

	result, err := repo.Create(data)

	if err != nil {
		// Convert CommonResponse to typed response
		errorResp := common.FromError(err)
		return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
			Status:    common.APIStatus.Invalid,
			Message:   errorResp.GetMessage(),
			ErrorCode: errorResp.GetErrorCode(),
		}
	}

	return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
		Status:  common.APIStatus.Ok,
		Data:    []*{{.EntityLower}}.{{.Entity}}{result},
		Message: "{{.Entity}} created successfully",
	}
}

// Get{{.Entity}}By{{.Entity}}ID retrieves a {{.EntityLower}} by its {{.Entity}}ID
func Get{{.Entity}}By{{.Entity}}ID({{.EntityLower}}ID string) *common.APIResponse[*{{.EntityLower}}.{{.Entity}}] {
	repo := {{.EntityLower}}.GetRepository()

	result, err := repo.GetBy{{.Entity}}ID({{.EntityLower}}ID)
	if err != nil {
		// Convert CommonResponse to typed response
		errorResp := common.FromError(err)
		return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
			Status:    errorResp.GetStatus(),
			Message:   errorResp.GetMessage(),
			ErrorCode: errorResp.GetErrorCode(),
		}
	}

	return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
		Status:  common.APIStatus.Ok,
		Data:    []*{{.EntityLower}}.{{.Entity}}{result},
		Message: "{{.Entity}} retrieved successfully",
	}
}

// List{{.EntityPlural}} retrieves a list of {{.EntityLower}}s with optional filtering
func List{{.EntityPlural}}(query *common.Query[{{.EntityLower}}.{{.Entity}}]) *common.APIResponse[*{{.EntityLower}}.{{.Entity}}] {
	repo := {{.EntityLower}}.GetRepository()

	filter := query.Filter
	offset := query.Offset
	limit := query.Limit
	sort := query.Sort

	if limit == 0 {
		limit = 10 // default limit
	}

	results, err := repo.List(filter, offset, limit, sort)
	if err != nil {
		// Convert CommonResponse to typed response
		errorResp := common.FromError(err)
		return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
			Status:    common.APIStatus.Invalid,
			Message:   errorResp.GetMessage(),
			ErrorCode: errorResp.GetErrorCode(),
		}
	}

	// Get total count for pagination
	total, err := repo.Count(filter)
	if err != nil {
		total = 0
	}

	return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
		Status:  common.APIStatus.Ok,
		Data:    results,
		Message: "{{.EntityPlural}} retrieved successfully",
		Total:   total,
	}
}

// Update{{.Entity}} updates an existing {{.EntityLower}}
func Update{{.Entity}}({{.EntityLower}}ID string, data *{{.EntityLower}}.{{.Entity}}) *common.APIResponse[*{{.EntityLower}}.{{.Entity}}] {
	repo := {{.EntityLower}}.GetRepository()
	result, err := repo.UpdateBy{{.Entity}}ID({{.EntityLower}}ID, data)
	if err != nil {
		// Convert CommonResponse to typed response
		errorResp := common.FromError(err)
		return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
			Status:    common.APIStatus.Invalid,
			Message:   errorResp.GetMessage(),
			ErrorCode: errorResp.GetErrorCode(),
		}
	}

	return &common.APIResponse[*{{.EntityLower}}.{{.Entity}}]{
		Status:  common.APIStatus.Ok,
		Data:    []*{{.EntityLower}}.{{.Entity}}{result},
		Message: "{{.Entity}} updated successfully",
	}
}

// Delete{{.Entity}} deletes a {{.EntityLower}} by ID (soft delete)
func Delete{{.Entity}}({{.EntityLower}}ID string) *common.APIResponse[any] {
	repo := {{.EntityLower}}.GetRepository()

	err := repo.DeleteBy{{.Entity}}ID({{.EntityLower}}ID)
	if err != nil {
		// Convert CommonResponse to typed response
		errorResp := common.FromError(err)
		return &common.APIResponse[any]{
			Status:    common.APIStatus.Invalid,
			Message:   errorResp.GetMessage(),
			ErrorCode: errorResp.GetErrorCode(),
		}
	}

	return &common.APIResponse[any]{
		Status:  common.APIStatus.Ok,
		Message: "{{.Entity}} deleted successfully",
	}
}
`

var API = `package api

import (
	{{if hasRequiredFields .Fields}}"regexp"
	"strings"{{end}}

	"gitlab.silvertiger.tech/go-sdk/go-common/common"
	"gitlab.silvertiger.tech/go-sdk/go-common/request"
	"gitlab.silvertiger.tech/go-sdk/go-common/responder"
	"{{.Module}}/internal/action"
	"{{.Module}}/{{.PkgPath}}"
	"{{.Module}}/constants"
)

{{if hasRequiredFields .Fields}}// Email validation regex
var emailRegex = regexp.MustCompile(` + "`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`" + `)

// isValidEmail validates email format
func isValidEmail(email string) bool {
	return emailRegex.MatchString(strings.TrimSpace(email))
}{{end}}

// Create{{.Entity}} creates a new {{.EntityLower}}
func Create{{.Entity}}(req request.APIRequest, res responder.APIResponder) error {
	var {{.EntityLower}}Data {{.Entity | lower}}.{{.Entity}}
	if err := req.ParseBody(&{{.EntityLower}}Data); err != nil {
		return res.Respond(common.NewErrorResponse(common.APIStatus.Invalid, "INVALID_REQUEST_BODY", "Failed to parse request body: "+err.Error()))
	}

{{generateValidation .Fields .EntityLower}}

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

{{generateValidation .Fields .EntityLower}}

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
