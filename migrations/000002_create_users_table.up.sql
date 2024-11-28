-- Create the 'users' table to store user information
CREATE TABLE IF NOT EXISTS users (
    id bigserial PRIMARY KEY, -- Unique identifier for each user
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), -- Timestamp when the user was created
    username text NOT NULL, -- Username for the user (required)
    email citext UNIQUE NOT NULL, -- User's email, case-insensitive and unique (required)
    password_hash bytea NOT NULL, -- Encrypted password hash (required)
    activated bool NOT NULL, -- Account activation status (true/false)
    version integer NOT NULL DEFAULT 1 -- Version for tracking record changes
);

-- Create the 'bookreviews' table to store reviews and ratings for books
CREATE TABLE IF NOT EXISTS bookreviews (
    id bigserial PRIMARY KEY, -- Unique identifier for each review
    book_id INT DEFAULT 0 REFERENCES books(id) ON DELETE CASCADE, -- Associated book, deleted if book is removed
    user_id INT REFERENCES users(id) ON DELETE CASCADE, -- Reviewer, deleted if user is removed
    rating FLOAT CHECK (rating BETWEEN 1 AND 5), -- Review rating (must be between 1 and 5)
    review TEXT, -- Written review content (optional)
    review_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP, -- Timestamp of the review creation
    version integer NOT NULL DEFAULT 1 -- Version for tracking record changes
);

-- Create the 'readinglists' table to store user-created reading lists
CREATE TABLE IF NOT EXISTS readinglists (
    id bigserial PRIMARY KEY, -- Unique identifier for each reading list
    name VARCHAR(255), -- Name of the reading list
    description TEXT, -- Description of the reading list
    created_by INT REFERENCES users(id) ON DELETE SET NULL, -- Creator of the reading list, set to NULL if user is removed
    version integer NOT NULL DEFAULT 1 -- Version for tracking record changes
);

-- Junction table for associating reading lists with books (many-to-many relationship)
CREATE TABLE IF NOT EXISTS readinglist_books (
    readinglist_id INT REFERENCES readinglists(id) ON DELETE CASCADE, -- Associated reading list, deleted if list is removed
    book_id INT REFERENCES books(id) ON DELETE CASCADE, -- Associated book, deleted if book is removed
    status VARCHAR(50) CHECK (status IN ('currently reading', 'completed')), -- Status of the book in the reading list
    version integer NOT NULL DEFAULT 1, -- Version for tracking record changes
    PRIMARY KEY (readinglist_id, book_id) -- Composite primary key for the junction table
);

-- Function to automatically calculate and update the average rating of a book
CREATE OR REPLACE FUNCTION automatic_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    -- Update the book's average rating based on associated reviews
    UPDATE books
    SET average_rating = (
        SELECT ROUND(CAST(AVG(rating) AS NUMERIC), 2) -- Calculate the average rating rounded to 2 decimal places
        FROM bookreviews
        WHERE bookreviews.book_id = NEW.book_id
    )
    WHERE id = NEW.book_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to execute the average rating update function after changes to 'bookreviews'
CREATE OR REPLACE TRIGGER update_book_rating
AFTER INSERT OR UPDATE OR DELETE ON bookreviews
FOR EACH ROW
EXECUTE FUNCTION automatic_average_rating();
