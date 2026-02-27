package models

import "gorm.io/gorm"

// Hackathon represents an IT event in the database.
type Hackathon struct {
	gorm.Model
	Title    string `json:"title" gorm:"not null"`
	Date     string `json:"date" gorm:"not null"`
	Format   string `json:"format" gorm:"not null"`
	City     string `json:"city"`
	AgeLimit string `json:"ageLimit"`
	Link     string `json:"link"`
}
