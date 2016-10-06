package models

import "github.com/jinzhu/gorm"

// User db Model
type User struct {
	gorm.Model
	Username string `gorm:"size:255"`
	Email    string `gorm:"type:varchar(100);unique_index"`
	Password string
}
