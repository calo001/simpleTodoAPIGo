package config

import "github.com/jinzhu/gorm"

var DB *gorm.DB

func Init() *gorm.DB {
	//db, err = gorm.Open("postgres", "host=localhost port=5432 user=admin dbname=tododb password=123  sslmode=disable")
	db, err := gorm.Open("postgres", "host=postgres.render.com port=5432 user=admin dbname=tododb password=tQGGa3UsRV")

	if err != nil {
		panic(err.Error())
	}

	DB = db
	return DB
}

func GetDB() *gorm.DB {
	return DB
}