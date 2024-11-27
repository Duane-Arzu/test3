CREATE TABLE products (
    product_id bigserial PRIMARY KEY,       -- Unique ID for each product
    name text NOT NULL,                     -- Product name
    description text NOT NULL,              -- Product description
    category text NOT NULL,                 -- Product category
    image_url text NOT NULL,                -- URL for the product's image
    price text NOT NULL,                    -- Product price as text
    avg_rating DECIMAL(3, 2) DEFAULT 0.00, -- Average rating based on reviews
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), -- Date product was added
    version integer NOT NULL DEFAULT 1      -- Version for tracking changes
);

-- Create a table to store reviews for products
CREATE TABLE reviews (
    review_id bigserial PRIMARY KEY,         -- Unique ID for each review
    product_id INT REFERENCES products(product_id) ON DELETE CASCADE, -- Links review to a product
    author VARCHAR(255),                     -- Name of the review author
    rating FLOAT CHECK (rating BETWEEN 1 AND 5), -- Review rating (1-5)
    comment text NOT NULL,               -- Review content
    helpful_count INT DEFAULT 0,             -- Count of helpful votes for the review
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(), -- Date review was created
    version integer NOT NULL DEFAULT 1       -- Version for tracking review updates
);

-- Function to automatically update a product's average rating when reviews change
CREATE OR REPLACE FUNCTION automatic_average_rating()
RETURNS TRIGGER AS $$
BEGIN
    -- Calculate and update the product's average rating when a new review is added or changed
    UPDATE products
    SET average_rating = (
        SELECT ROUND(CAST(AVG(rating) AS NUMERIC), 2)  -- Calculates the average rating, rounded to 2 decimals
        FROM reviews
        WHERE reviews.product_id = NEW.product_id
    )
    WHERE product_id = NEW.product_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to run automatic_average_rating() after reviews are added, updated, or deleted
CREATE OR REPLACE TRIGGER update_product_rating
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW
EXECUTE FUNCTION automatic_average_rating();