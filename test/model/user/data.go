package user

import "time"

// @entity db:users
// @index name:text
// @index email:1,is_active:1
type User struct {
	ID        string    `json:"id" bson:"_id" validate:"required"`
	UserID    string    `json:"user_id" bson:"user_id" validate:"required,alphanum" index:"unique"`
	Name      string    `json:"name" bson:"name" validate:"required,min=2,max=100" index:"text"`
	Email     string    `json:"email" bson:"email" validate:"required,email" index:"1"`
	IsActive  bool      `json:"is_active" bson:"is_active" index:"1"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}
