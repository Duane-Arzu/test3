// Filename: internal/data/reviews.go
package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Duane-Arzu/test3/internal/validator"
)

// Review struct represents a review for a product, with various attributes related to the review's content and metadata.
type Review struct {
	ReviewID     int64     `json:"review_id"`     // Unique identifier for the review (primary key)
	ProductID    int64     `json:"product_id"`    // Identifier of the product being reviewed (foreign key)
	Author       string    `json:"author"`        // Name of the review's author
	Rating       int64     `json:"rating"`        // Rating given by the author, constrained to values between 1 and 5
	Comment      string    `json:"commentt"`      // Content of the comment, required field
	HelpfulCount int32     `json:"helpful_count"` // Number of "helpful" votes, defaults to 0 if not specified
	CreatedAt    time.Time `json:"-"`             // Timestamp for when the review was created, auto-set to current time
	Version      int       `json:"version"`       // Version number to track changes to the review
}

// ReviewModel wraps the database connection pool for managing review data.
type ReviewModel struct {
	DB *sql.DB // Database connection pool
}

// ValidateReview validates required fields and checks constraints on a Review struct.
func ValidateReview(v *validator.Validator, review *Review) {
	v.Check(review.Author != "", "author", "must be provided")                             // Ensures author field is not empty
	v.Check(review.Comment != "", "comment", "must be provided")                           // Ensures review_text is provided
	v.Check(len(review.Author) <= 25, "author", "must not be more than 25 bytes long")     // Restricts author length to 25 bytes
	v.Check(review.ProductID > 0, "product_id", "must be a positive integer")              // ProductID must be a valid positive integer
	v.Check(review.Rating >= 1 && review.Rating <= 5, "rating", "must be between 1 and 5") // Rating must be between 1 and 5
}

// InsertReview adds a new review to the database and retrieves its ID, creation timestamp, and version.
func (c ReviewModel) InsertReview(review *Review) error {
	query := `
		INSERT INTO reviews (product_id, author, rating, comment, helpful_count)
		VALUES ($1, $2, $3, $4, COALESCE($5, 0))
		RETURNING review_id, created_at, version
	`
	args := []any{review.ProductID, review.Author, review.Rating, review.Comment, review.HelpfulCount}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // Ensure the timeout context is canceled to free up resources

	// Execute query and store the new review's ID, creation timestamp, and version
	return c.DB.QueryRowContext(ctx, query, args...).Scan(
		&review.ReviewID,
		&review.CreatedAt,
		&review.Version)
}

// GetReview retrieves a single review by its ID. Returns ErrRecordNotFound if no review is found.
func (c ReviewModel) GetReview(id int64) (*Review, error) {
	if id < 1 {
		return nil, ErrRecordNotFound // Validates ID input to avoid invalid queries
	}
	query := `
		SELECT review_id, product_id, author, rating, comment, helpful_count, created_at, version
		FROM reviews
		WHERE review_id = $1
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Execute query to fetch review details
	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ReviewID,
		&review.ProductID,
		&review.Author,
		&review.Rating,
		&review.Comment,
		&review.HelpfulCount,
		&review.CreatedAt,
		&review.Version,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &review, nil
}

// UpdateReview modifies an existing review's details and increments its version number.
func (c ReviewModel) UpdateReview(review *Review) error {
	query := `
		UPDATE reviews
		SET author = $1, rating = $2, comment = $3, version = version + 1
		WHERE review_id = $4
		RETURNING version
	`
	args := []any{review.Author, review.Rating, review.Comment, review.ReviewID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, query, args...).Scan(&review.Version) // Update version for tracking changes
}

// DeleteReview removes a review from the database by ID.
func (c ReviewModel) DeleteReview(id int64) error {
	if id < 1 {
		return ErrRecordNotFound // Validate ID to prevent unnecessary database operations
	}
	query := `
		DELETE FROM reviews
		WHERE review_id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := c.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// GetAllReviews retrieves a list of reviews matching a given author name with sorting and pagination.
func (c ReviewModel) GetAllReviews(author string, filters Filters) ([]*Review, Metadata, error) {
	query := fmt.Sprintf(`
	SELECT COUNT(*) OVER(), review_id, product_id, author, rating, comment, helpful_count, created_at, version
	FROM reviews
	WHERE (to_tsvector('simple', author) @@ plainto_tsquery('simple', $1) OR $1 = '') 
	ORDER BY %s %s, review_id ASC 
	LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, author, filters.limit(), filters.offset())
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	var totalRecords int
	reviews := []*Review{}

	// Process each row and populate reviews slice
	for rows.Next() {
		var review Review
		if err := rows.Scan(&totalRecords, &review.ReviewID, &review.ProductID, &review.Author, &review.Rating, &review.Comment, &review.HelpfulCount, &review.CreatedAt, &review.Version); err != nil {
			return nil, Metadata{}, err
		}
		reviews = append(reviews, &review)
	}

	// Check for any row iteration errors
	if err := rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// Calculate pagination metadata
	metadata := calculateMetaData(totalRecords, filters.Page, filters.PageSize)

	return reviews, metadata, nil
}

// GetAllProductReviews fetches all reviews associated with a specified product ID.
func (c ReviewModel) GetAllProductReviews(productID int64) ([]Review, error) {
	if productID < 1 {
		return nil, ErrRecordNotFound // Validate product ID before querying
	}

	query := `
		SELECT review_id, author, rating, comment, helpful_count, created_at, version
		FROM reviews
		WHERE product_id = $1
	`
	var reviews []Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := c.DB.QueryContext(ctx, query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan each row and populate reviews slice
	for rows.Next() {
		var review Review
		err := rows.Scan(
			&review.ReviewID,
			&review.Author,
			&review.Rating,
			&review.Comment,
			&review.HelpfulCount,
			&review.CreatedAt,
			&review.Version,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	// Check for errors after row iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return reviews, nil
}

// UpdateHelpfulCount increments the helpful_count for a review by 1.
func (c *ReviewModel) UpdateHelpfulCount(id int64) (*Review, error) {
	query := `
        UPDATE reviews
        SET helpful_count = helpful_count + 1
        WHERE review_id = $1
        RETURNING review_id, author, rating, comment, helpful_count, version
    `
	var review Review
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Update the helpful count and retrieve updated review fields
	err := c.DB.QueryRowContext(ctx, query, id).Scan(
		&review.ReviewID,
		&review.Author,
		&review.Rating,
		&review.Comment,
		&review.HelpfulCount,
		&review.Version,
	)
	if err != nil {
		return nil, err
	}

	return &review, nil
}

func (m *ProductModel) ProductExists(productID int64) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM products WHERE product_id = $1)`
	var exists bool
	err := m.DB.QueryRow(query, productID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
func (m *ReviewModel) Exists(id int64) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM reviews WHERE review_id = $1)`
	err := m.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (c ReviewModel) GetProductReview(rid int64, pid int64) (*Review, error) {
	//validate id
	if pid < 1 || rid < 1 {
		return nil, ErrRecordNotFound
	}

	//query
	query := `SELECT review_id, product_id, author, rating, comment, helpful_count, created_at, version
	FROM reviews
	WHERE review_id = $1 AND product_id = $2
	`
	var review Review

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := c.DB.QueryRowContext(ctx, query, rid, pid).Scan(
		&review.ReviewID,
		&review.ProductID,
		&review.Author,
		&review.Rating,
		&review.Comment,
		&review.HelpfulCount,
		&review.CreatedAt,
		&review.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &review, nil
}
