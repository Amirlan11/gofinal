package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Tokens TokenModel
	Users  UserModel
	Foods  FoodModel
	Sales  SaleModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens: TokenModel{DB: db},
		Users:  UserModel{DB: db},
		Foods:  FoodModel{DB: db},
		Sales:  SaleModel{DB: db},
	}
}
