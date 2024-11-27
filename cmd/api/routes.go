package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func (a *applicationDependencies) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(a.notFoundResponse)

	router.MethodNotAllowed = http.HandlerFunc(a.methodNotAllowedResponse)

	//Product part
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", a.healthcheckHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product", a.listProductHandler)
	router.HandlerFunc(http.MethodPost, "/v1/product", a.createProductHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid", a.displayProductHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/product/:pid", a.updateProductHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/product/:pid", a.deleteProductHandler)

	// //Review part
	router.HandlerFunc(http.MethodGet, "/v1/review", a.listReviewHandler)
	router.HandlerFunc(http.MethodPost, "/v1/review", a.createReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/review/:rid", a.displayReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/review/:rid", a.updateReviewHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/review/:rid", a.deleteReviewHandler)

	router.HandlerFunc(http.MethodGet, "/v1/product-review/:rid", a.listProductReviewHandler)
	router.HandlerFunc(http.MethodGet, "/v1/product/:pid/review/:rid", a.getProductReviewHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/helpful-count/:rid", a.HelpfulCountHandler)

	return a.recoverPanic(a.rateLimit(router))

}
