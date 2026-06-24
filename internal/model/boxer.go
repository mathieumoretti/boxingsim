package model

import "time"

// Boxer represents a boxer in the system
type Boxer struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Name         string    `json:"name"`
	Nickname     *string   `json:"nickname"`
	PositionX    float64   `json:"position_x"`
	PositionY    float64   `json:"position_y"`
	Health       float64   `json:"health"`
	Energy       float64   `json:"energy"`
	Strength     float64   `json:"strength"`
	Defense      float64   `json:"defense"`
	Agility      float64   `json:"agility"`
	Experience   float64   `json:"experience"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// BoxerCreate represents a request to create a new boxer
type BoxerCreate struct {
	Name         string  `json:"name" binding:"required,min=1,max=100"`
	Nickname     *string `json:"nickname"`
	PositionX    float64 `json:"position_x" binding:"min=0"`
	PositionY    float64 `json:"position_y" binding:"min=0"`
	Strength     float64 `json:"strength" binding:"min=0"`
	Defense      float64 `json:"defense" binding:"min=0"`
	Agility      float64 `json:"agility" binding:"min=0"`
}

// BoxerUpdate represents a request to update a boxer
type BoxerUpdate struct {
	Name         *string `json:"name" binding:"omitempty,min=1,max=100"`
	Nickname     *string `json:"nickname"`
	PositionX    *float64 `json:"position_x" binding:"omitempty,min=0"`
	PositionY    *float64 `json:"position_y" binding:"omitempty,min=0"`
	Strength     *float64 `json:"strength" binding:"omitempty,min=0"`
	Defense      *float64 `json:"defense" binding:"omitempty,min=0"`
	Agility      *float64 `json:"agility" binding:"omitempty,min=0"`
}

// BoxerResponse represents a boxer for API responses
type BoxerResponse struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Name         string    `json:"name"`
	Nickname     *string   `json:"nickname"`
	PositionX    float64   `json:"position_x"`
	PositionY    float64   `json:"position_y"`
	Health       float64   `json:"health"`
	Energy       float64   `json:"energy"`
	Strength     float64   `json:"strength"`
	Defense      float64   `json:"defense"`
	Agility      float64   `json:"agility"`
	Experience   float64   `json:"experience"`
	Level        int       `json:"level"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StatsAdd represents stats to add to a boxer
type StatsAdd struct {
	Strength    *float64 `json:"strength"`
	Defense     *float64 `json:"defense"`
	Agility     *float64 `json:"agility"`
	Experience  *float64 `json:"experience"`
}

// UserCreate represents a request to create a user and boxer
type UserCreate struct {
	ID           int    `json:"id"`
	Username     string `json:"username" binding:"required,min=1,max=100"`
	Email        string `json:"email" binding:"required,email"`
	Password     string `json:"password" binding:"required,min=6"`
	BoxerName    string `json:"boxer_name" binding:"required,min=1,max=100"`
}