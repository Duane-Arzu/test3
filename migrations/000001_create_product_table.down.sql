-- First, drop the trigger associated with updating the product rating
-- This is necessary because the trigger relies on the function and the "reviews" table
DROP TRIGGER IF EXISTS update_product_rating ON reviews;

-- Next, remove the function for calculating average ratings
-- This function should be dropped before deleting the tables it interacts with
DROP FUNCTION IF EXISTS automatic_average_rating();

-- Drop the tables in order, beginning with the one that has dependencies on another table
-- Dropping "reviews" first, as it references "products" via foreign keys
DROP TABLE IF EXISTS reviews;

-- Finally, drop the "products" table to complete the cleanup
DROP TABLE IF EXISTS products;
