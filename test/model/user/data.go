package user

import "time"

// @entity db:users
// @index email:1 unique
// @index name:1,created_at:-1
// @index email:text
type User struct {
	ID        string    `json:"id" bson:"_id" validate:"required"`
	Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100" index:"1"`
	Email     string    `json:"email" bson:"email" validate:"required,email" index:"unique"`
	Age       int       `json:"age" bson:"age" validate:"min=0,max=150"`
	IsActive  bool      `json:"is_active" bson:"is_active" index:"1"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
