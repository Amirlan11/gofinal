package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/foods", app.listFoodsHandler)
	router.HandlerFunc(http.MethodPost, "/v1/foods", app.requireActivatedUser(app.createFoodHandler))
	router.HandlerFunc(http.MethodGet, "/v1/foods/:id", app.showFoodHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/foods/:id", app.requireActivatedUser(app.updateFoodHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/foods/:id", app.requireActivatedUser(app.deleteFoodHandler))

	router.HandlerFunc(http.MethodGet, "/v1/sales", app.listSalesHandler)
	router.HandlerFunc(http.MethodPost, "/v1/sales", app.requireActivatedUser(app.createSaleHandler))
	router.HandlerFunc(http.MethodGet, "/v1/sales/:id", app.showSaleHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/sales/:id", app.requireActivatedUser(app.updateSaleHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/sales/:id", app.requireActivatedUser(app.deleteSaleHandler))

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/upload", app.ImageHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
