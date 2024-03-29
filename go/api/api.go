package api

import (
	"database/sql"
	"encoding/json"
	"html"
	"net/http"
	"strings"

	// "time"

	"github.com/jakubruminski/FYP/go/api/fetch"
	"github.com/jakubruminski/FYP/go/api/product"
	"github.com/jakubruminski/FYP/go/api/query"

	"github.com/jakubruminski/FYP/go/utils/http/response"
	"github.com/jakubruminski/FYP/go/utils/logger"
	"github.com/jakubruminski/FYP/go/utils/postgres"
	"github.com/jakubruminski/FYP/go/utils/token"
)

type Products struct {
	Results  *[]*product.Product               `json:"results"`
	Currency map[string]map[string]interface{} `json:"currency"`
}

func GetResponse(logger *logger.Logger, r *http.Request, w http.ResponseWriter) (jsonResponse []byte, ok bool) {
	// for {
	// 	// sleep for 1 second
	// 	time.Sleep(1 * time.Second)
	// }

	if r.URL.Path == "/api/search" {
		return getProductsHandler(logger, w, r)

	} else if r.URL.Path == "/api/add_item" {
		return addItemHandler(logger, w, r)
	}

	logger.ERROR("Invalid request %s", r.URL.Path)
	return nil, false
}

func getProductsHandler(logger *logger.Logger, w http.ResponseWriter, r *http.Request) (jsonResponse []byte, ok bool) {
	searchTerm := parseSearchValue(r.FormValue("search_term"))
	searchTerm = strings.ToLower(searchTerm)

	products := &[]*product.Product{}

	ok = postgres.ExecuteInTransaction(logger, getProducts_DoInTransaction, products, searchTerm)
	if !ok && len(*products) == 0 {
		logger.ERROR("Failed to get products")
		return nil, false
	}

	currency, ok := getCurrency(logger)
	if !ok {
		logger.ERROR("Failed to get currency")
		return nil, false
	}

	jsonResponse, err := json.Marshal(Products{Results: products, Currency: currency})
	if err != nil {
		logger.ERROR("Failed to marshal response")
		return nil, false
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)

	return jsonResponse, true
}


func getProducts_DoInTransaction(logger *logger.Logger, tx *sql.Tx, args ...interface{}) bool {

    if len(args) != 2 {
        logger.ERROR("Expected 2 arguments, got %d", len(args))
        return false
    }

    products, ok := args[0].(*[]*product.Product)
    if !ok {
        logger.ERROR("Failed to get products")
        return false
    }

    searchTerm, ok := args[1].(string)
    if !ok {
		logger.ERROR("Failed to get search term")
        return false
    }

	found, ok := query.Products(logger, tx, products, searchTerm)
	if !ok {
		logger.ERROR("Failed to get products from database")
	}
	if found {
		logger.INFO("Products found in database")
		return true
	}

	logger.ERROR("No products matched in database, now web scraping...")
	ok = fetch.Products(logger, products, searchTerm)
	if !ok {
		logger.ERROR("Failed to get products from web scraping")
		return false
	}

	if len(*products) == 0 {
		logger.ERROR("No products found")
		return true
	}

	ok = query.AddProducts(logger, tx, searchTerm, products)
	if !ok {
		logger.ERROR("Failed to add products to database")
		return false
	}

	ok = query.AddSearchTerm(logger, tx, searchTerm, products)
	if !ok {
		logger.ERROR("Failed to add search term to database")
		return false
	}

	return true
}


// TODO: These should be fetched and not hardcoded.
func getCurrency(logger *logger.Logger) (Rates map[string]map[string]interface{}, ok bool) {
	Rates = map[string]map[string]interface{}{
		"Canada":     {"rate": 1.44, "symbol": "C$"},
		"India":      {"rate": 89.42, "symbol": "₹"},
		"Costa Rica": {"rate": 588.03, "symbol": "₡"},
		"Australia":  {"rate": 1.63, "symbol": "A$"},
		"UK":         {"rate": 0.86, "symbol": "£"},
		"Euro":       {"rate": 1.0, "symbol": "€"},
		"Poland":     {"rate": 4.46, "symbol": "zł"},
	}

	return Rates, true
}

// This function escapes html characters and replaces spaces with "%20"
func parseSearchValue(searchValue string) string {

	startQuote := ""
	endQuote := ""
	if strings.HasPrefix(searchValue, "\"") && strings.HasSuffix(searchValue, "\"") {
		searchValue = strings.Trim(searchValue, "\"")
		startQuote = "\""
		endQuote = "\""
	}

	searchValue = html.EscapeString(searchValue)

	// if there is a space, replace it with "%20"
	searchValue = strings.Replace(searchValue, " ", "%20", -1)

	searchValue = startQuote + searchValue + endQuote

	return searchValue
}

func addItemHandler(logger *logger.Logger, w http.ResponseWriter, r *http.Request) (jsonResponse []byte, ok bool) {
	logger.INFO("Request: %s", r.URL.Path)

	clientID, ok := token.GetID(logger, r)
	if !ok {
		response.WriteResponse(logger, w, http.StatusUnauthorized, "application/json", "error", "Unauthorized")
		return nil, false
	}

	product, ok := product.ParseProduct(logger, r)
	if !ok {
		response.WriteResponse(logger, w, http.StatusBadRequest, "application/json", "error", "Failed to parse product")
		return nil, false
	}

	ok = query.AddToBaskets(logger, clientID, *product)
	if !ok {
		response.WriteResponse(logger, w, http.StatusInternalServerError, "application/json", "error", "Failed to add item")
		return nil, false
	}

	return nil, true
}
