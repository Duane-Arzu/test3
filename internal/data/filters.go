// Filename: internal/data/filters.go
package data

import (
	"strings"

	"github.com/Duane-Arzu/test3/internal/validator"
	_ "github.com/Duane-Arzu/test3/internal/validator"
)

// The Filters struct holds pagination and sorting parameters
// that help in managing paginated results for client requests.
type Filters struct {
	Page         int      // Specifies the page number requested by the client.
	PageSize     int      // Specifies the number of records per page.
	Sort         string   // Field by which to sort the results, with optional direction.
	SortSafeList []string // List of allowed fields for sorting to prevent unsafe queries.
}

// The Metadata struct contains pagination details
// that will be sent back to the client.
type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`  // Indicates the current page in the paginated results.
	PageSize     int `json:"page_size,omitempty"`     // Specifies the number of items per page.
	FirstPage    int `json:"first_page,omitempty"`    // The first page in the dataset (usually 1).
	LastPage     int `json:"last_page,omitempty"`     // The last available page based on total records.
	TotalRecords int `json:"total_records,omitempty"` // The total count of records across all pages.
}

// ValidateFilters checks that the pagination and sorting parameters
// in Filters struct are valid and within acceptable ranges.
func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")             // Ensure page number is positive.
	v.Check(f.Page <= 500, "page", "must be a maximum of 500")           // Limit page number to a maximum of 500.
	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")    // Ensure page size is positive.
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")  // Limit page size to a maximum of 100 records.
	v.Check(validator.PermittedValue(f.Sort, f.SortSafeList...), "sort", // Validate sort field is allowed.
		"invalid sort value")
}

// sortColumn returns the sanitized sort field by removing any
// direction indicator (like '-') to prevent SQL injection risks.
func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-") // Remove prefix for consistency.
		}
	}
	// Prevent operation if unsafe sort parameter detected,
	// which could be used for SQL injection.
	panic("unsafe sort parameter: " + f.Sort)
}

// sortDirection determines the direction of sorting
// (ASC for ascending, DESC for descending) based on the prefix.
func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC" // Indicates descending order.
	}
	return "ASC" // Default to ascending order.
}

// limit returns the page size, representing the number of records per page.
func (f Filters) limit() int {
	return f.PageSize
}

// offset calculates the starting position of records to skip,
// based on the current page, for pagination purposes.
func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

// calculateMetaData generates pagination metadata based on the total
// number of records, current page, and page size, making it easier for
// the client to understand paginated navigation.
func calculateMetaData(totalRecords int, currentPage int, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{} // Return empty metadata if there are no records.
	}

	return Metadata{
		CurrentPage:  currentPage,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     (totalRecords + pageSize - 1) / pageSize, // Calculate the last page.
		TotalRecords: totalRecords,
	}
}
