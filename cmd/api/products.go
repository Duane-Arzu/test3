package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Duane-Arzu/test3/internal/data"
	_ "github.com/Duane-Arzu/test3/internal/data"
	"github.com/Duane-Arzu/test3/internal/validator"
	_ "github.com/Duane-Arzu/test3/internal/validator"
)

// Product represents the expected structure for incoming product data with optional fields
// Using pointers allows us to distinguish between empty values and omitted fields in PATCH requests
type incomingProductData struct {
	Name        *string  `json:"name"`
	Description *string  `json:"description"`
	Category    *string  `json:"category"`
	ImageURL    *string  `json:"image_url"`
	Price       *string  `json:"price"`
	AvgRating   *float32 `json:"avg_rating"`
}

// createProductHandler handles POST requests to create new products
// It validates the incoming data and returns the created product with its ID
func (a *applicationDependencies) createProductHandler(w http.ResponseWriter, r *http.Request) {
	// Define structure for incoming product creation data
	// Note: All fields are required for creation
	var incomingProductData struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Category    string `json:"category"`
		ImageURL    string `json:"image_url"`
		Price       string `json:"price"`
	}

	// Parse JSON request body into our data structure
	err := a.readJSON(w, r, &incomingProductData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Create new product instance from incoming data
	product := &data.Product{
		Name:        incomingProductData.Name,
		Description: incomingProductData.Description,
		Category:    incomingProductData.Category,
		ImageURL:    incomingProductData.ImageURL,
		Price:       incomingProductData.Price,
	}

	// Validate the product data using our validation package
	v := validator.New()
	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the validated product into the database
	err = a.productModel.InsertProduct(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Set the Location header to point to the newly created product
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("products/%d", product.ProductID))

	// Return the created product in the response
	data := envelope{
		"Product": product,
	}
	err = a.writeJSON(w, http.StatusCreated, data, headers)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// displayProductHandler handles GET requests for retrieving a single product by ID
// Returns 404 if the product doesn't exist
func (a *applicationDependencies) displayProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the product ID from the URL parameters
	id, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Attempt to retrieve the product from the database
	product, err := a.productModel.GetProduct(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.notFoundResponse(w, r)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return the found product in the response
	data := envelope{
		"Product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// updateProductHandler handles PATCH requests to update existing products
// Supports partial updates using pointer fields to distinguish between zero values and omitted fields
func (a *applicationDependencies) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the product ID from the URL parameters
	id, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Retrieve the existing product from the database
	product, err := a.productModel.GetProduct(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			a.notFoundResponse(w, r)
		} else {
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Define structure for partial updates using pointer fields
	var incomingProductData struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Category    *string `json:"category"`
		ImageURL    *string `json:"image_url"`
		Price       *string `json:"price"`
		// Commented fields can be uncommented when needed
		//UpdatedAt   *time.Time `json:"updated_at"`
		//AvgRating *float64   `json:"avg_rating"`
	}

	// Parse the JSON request body
	err = a.readJSON(w, r, &incomingProductData)
	if err != nil {
		a.badRequestResponse(w, r, err)
		return
	}

	// Update only the fields that were provided in the request
	if incomingProductData.Name != nil {
		product.Name = *incomingProductData.Name
	}
	if incomingProductData.Description != nil {
		product.Description = *incomingProductData.Description
	}
	if incomingProductData.Category != nil {
		product.Category = *incomingProductData.Category
	}
	if incomingProductData.ImageURL != nil {
		product.ImageURL = *incomingProductData.ImageURL
	}
	if incomingProductData.Price != nil {
		product.Price = *incomingProductData.Price
	}

	// Validate the updated product data
	v := validator.New()
	data.ValidateProduct(v, product)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Save the updated product to the database
	err = a.productModel.UpdateProduct(product)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return the updated product in the response
	data := envelope{
		"Product": product,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// deleteProductHandler handles DELETE requests to remove products from the system
// Returns a success message if the product was successfully deleted
func (a *applicationDependencies) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	// Extract and validate the product ID from the URL parameters
	id, err := a.readIDParam(r, "pid")
	if err != nil {
		a.notFoundResponse(w, r)
		return
	}

	// Attempt to delete the product from the database
	err = a.productModel.DeleteProduct(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			a.PIDnotFound(w, r, id)
		default:
			a.serverErrorResponse(w, r, err)
		}
		return
	}

	// Return a success message
	data := envelope{
		"message": "Product successfully deleted",
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}

// listProductHandler handles GET requests to retrieve a filtered, paginated list of products
// Supports filtering by name and category, with sorting and pagination options
func (a *applicationDependencies) listProductHandler(w http.ResponseWriter, r *http.Request) {
	// Define structure to hold query parameters and filtering options
	var queryParametersData struct {
		Name     string
		Category string
		data.Filters
	}

	// Extract query parameters from the URL
	queryParameters := r.URL.Query()
	queryParametersData.Name = a.getSingleQueryParameter(queryParameters, "name", "")
	queryParametersData.Category = a.getSingleQueryParameter(queryParameters, "category", "")

	// Set up and validate pagination and sorting parameters
	v := validator.New()
	queryParametersData.Filters.Page = a.getSingleIntegerParameter(queryParameters, "page", 1, v)
	queryParametersData.Filters.PageSize = a.getSingleIntegerParameter(queryParameters, "page_size", 10, v)
	queryParametersData.Filters.Sort = a.getSingleQueryParameter(queryParameters, "sort", "product_id")
	queryParametersData.Filters.SortSafeList = []string{"product_id", "name", "-product_id", "-name"}

	// Validate the filters
	data.ValidateFilters(v, queryParametersData.Filters)
	if !v.IsEmpty() {
		a.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve filtered and paginated products from the database
	products, metadata, err := a.productModel.GetAllProducts(
		queryParametersData.Name,
		queryParametersData.Category,
		queryParametersData.Filters,
	)
	if err != nil {
		a.serverErrorResponse(w, r, err)
		return
	}

	// Return the products and metadata in the response
	data := envelope{
		"products":  products,
		"@metadata": metadata,
	}
	err = a.writeJSON(w, http.StatusOK, data, nil)
	if err != nil {
		a.serverErrorResponse(w, r, err)
	}
}
