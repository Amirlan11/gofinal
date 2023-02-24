package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	data2 "github.com/laldil/greenlight/internal/data"
	"github.com/laldil/greenlight/internal/validator"
)

func (app *application) createFoodHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title    string   `json:"title"`
		Price    int32    `json:"price"`
		Waittime int32    `json:"waittime"`
		Recipe   []string `json:"recipe"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	food := &data2.Food{
		Title:    input.Title,
		Price:    input.Price,
		Waittime: input.Waittime,
		Recipe:   input.Recipe,
	}

	v := validator.New()

	if data2.ValidateFood(v, food); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Foods.Insert(food)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/foods/%d", food.ID))

	err = app.writeJSON(w, http.StatusCreated, envelope{"food": food}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) showFoodHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	food, err := app.models.Foods.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"food": food}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateFoodHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	food, err := app.models.Foods.Get(id)
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
		Title    *string  `json:"title"`
		Price    *int32   `json:"price"`
		Waittime *int32   `json:"waittime"`
		Recipe   []string `json:"recipe"`
	}

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	if input.Title != nil {
		food.Title = *input.Title
	}

	if input.Price != nil {
		food.Price = *input.Price
	}

	if input.Waittime != nil {
		food.Waittime = *input.Waittime
	}

	if input.Recipe != nil {
		food.Recipe = input.Recipe
	}

	v := validator.New()

	if data2.ValidateFood(v, food); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.models.Foods.Update(food)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"food": food}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteFoodHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Foods.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data2.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "food successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) listFoodsHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Recipe []string
		data2.Filters
	}

	v := validator.New()
	qs := r.URL.Query()

	input.Title = app.readString(qs, "title", "")
	input.Recipe = app.readCSV(qs, "recipe", []string{})
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)
	input.Filters.Sort = app.readString(qs, "sort", "id")
	input.Filters.SortSafeList = []string{"id", "title", "price", "waittime", "-id", "-title", "-price", "-waittime"}

	if data2.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	foods, metadata, err := app.models.Foods.GetAll(input.Title, input.Recipe, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"foods": foods, "metadata": metadata}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) ImageHandler(w http.ResponseWriter, r *http.Request) {

	r.ParseMultipartForm(10 * 1024 * 1024)

	file, handler, err := r.FormFile("myfile")

	if err != nil {
		fmt.Println(err)
		return
	}

	defer file.Close()

	fmt.Println("File info")
	fmt.Println("File name:", handler.Filename)
	fmt.Println("File size:", handler.Size)
	fmt.Println("File type:", handler.Header.Get("Content-Type"))

	tempFile, err2 := ioutil.TempFile("images", "upload-*.jpg")
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	defer tempFile.Close()

	tempFile, err4 := ioutil.TempFile("pdfs", "upload-*.pdf")
	if err2 != nil {
		fmt.Println(err4)
		return
	}
	defer tempFile.Close()

	fileBytes, err3 := ioutil.ReadAll(file)
	if err3 != nil {
		fmt.Println(err2)
		return
	}
	tempFile.Write(fileBytes)
	fmt.Println("File uploaded")

}
