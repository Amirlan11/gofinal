package main

import (
	"errors"
	"fmt"
	data2 "github.com/laldil/greenlight/internal/data"
	"github.com/laldil/greenlight/internal/validator"
	"net/http"
)

func (app *application) createSaleHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title       string        `json:"title"`
		Description string        `json:"description"`
		Duration    data2.Runtime `json:"duration"`
		Foodsale    []string      `json:"foodsale"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	sale := &data2.Sale{
		Title:       input.Title,
		Description: input.Description,
		Duration:    input.Duration,
		Foodsale:    input.Foodsale,
	}

	v := validator.New()

	if data2.ValidateSale(v, sale); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Sales.Insert(sale)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/sales/%d", sale.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"sale": sale}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showSaleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sale, err := app.models.Sales.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sale": sale}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateSaleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	sale, err := app.models.Sales.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrInvalidRuntimeFormat):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	var input struct {
		Title       *string        `json:"title"`
		Description *string        `json:"description"`
		Duration    *data2.Runtime `json:"duration"`
		Foodsale    []string       `json:"foodsale"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		sale.Title = *input.Title
	}

	if input.Description != nil {
		sale.Description = *input.Description
	}

	if input.Duration != nil {
		sale.Duration = *input.Duration
	}

	if input.Foodsale != nil {
		sale.Foodsale = input.Foodsale
	}

	v := validator.New()

	if data2.ValidateSale(v, sale); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Sales.Update(sale)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sale": sale}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteSaleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Sales.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "sale successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listSalesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string
		Foodsale []string
		data2.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Foodsale = app.readCSV(qs, "foodsale", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "description", "duration", "-id", "-title", "-description", "-duration"}

	if data2.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	sales, metadata, err := app.models.Sales.GetAll(input.Title, input.Foodsale, input.Filters)
	fmt.Println(sales[1].Title)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"sales": sales, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
