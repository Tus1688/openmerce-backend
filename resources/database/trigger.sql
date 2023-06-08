DELIMITER //

CREATE TRIGGER update_cumulative_review AFTER INSERT ON reviews
    FOR EACH ROW
BEGIN
    DECLARE avg_review DECIMAL(2,1);

    -- Calculate the average review for the product
    SELECT AVG(rating) INTO avg_review FROM reviews WHERE product_refer = NEW.product_refer;

    -- Update the cumulative review in the products table
    UPDATE products SET cumulative_review = avg_review WHERE id = NEW.product_refer;
END//

DELIMITER ;
