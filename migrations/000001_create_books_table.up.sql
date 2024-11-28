-- Create the 'books' table to store book information
CREATE TABLE books (
    id bigserial PRIMARY KEY, -- Unique identifier for each book (auto-incrementing)
    title VARCHAR(255) NOT NULL, -- Title of the book (required)
    authors TEXT, -- Authors of the book (optional, allows multiple authors as text)
    isbn VARCHAR(20) NOT NULL, -- ISBN number of the book (required, up to 20 characters)
    publication_date TEXT, -- Publication date of the book (optional, stored as text)
    genre VARCHAR(100), -- Genre or category of the book (optional)
    description TEXT, -- Short description or summary of the book (optional)
    average_rating DECIMAL(3, 2) DEFAULT 0.00, -- Average user rating, defaults to 0.00
    version integer NOT NULL DEFAULT 1 -- Version for tracking updates to the book record
);
