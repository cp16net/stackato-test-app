package models

import "github.com/jinzhu/gorm"

// User db Model
type User struct {
	gorm.Model
	Username string `gorm:"size:255"`
	Emails   []Email
	Password string
}

// Email db Model
type Email struct {
	ID     int
	UserID int    `gorm:"index"`
	Email  string `gorm:"type:varchar(100);unique_index"`
}
