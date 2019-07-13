package model

import "github.com/jinzhu/gorm"

type User struct {
	gorm.Model
	Username string `json:"username"`
	Password string `json:"password"`
	Todos    []Task `json:"todos"`
}

type Task struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	UserID      uint   `json:"userid"`
}
