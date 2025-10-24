INSERT INTO categories (id, name, description) VALUES
    (1, 'Beverages', 'Soft drinks, coffees, teas, beers, and ales'),
    (2, 'Condiments', 'Sweet and savory sauces, relishes, spreads, and seasonings'),
    (3, 'Produce', 'Dried fruit and bean curd')
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    description = EXCLUDED.description;

INSERT INTO suppliers (id, company_name, contact_name, contact_title, address, city, region, postal_code, country, phone) VALUES
    (1, 'Exotic Liquids', 'Charlotte Cooper', 'Purchasing Manager', '49 Gilbert St.', 'London', NULL, 'EC1 4SD', 'UK', '(171) 555-2222'),
    (2, 'New Orleans Cajun Delights', 'Shelley Burke', 'Order Administrator', 'P.O. Box 78934', 'New Orleans', 'LA', '70117', 'USA', '(100) 555-4822'),
    (3, 'Grandma Kelly''s Homestead', 'Regina Murphy', 'Sales Representative', '707 Oxford Rd.', 'Ann Arbor', 'MI', '48104', 'USA', '(313) 555-5735')
ON CONFLICT (id) DO UPDATE SET
    company_name = EXCLUDED.company_name,
    contact_name = EXCLUDED.contact_name,
    contact_title = EXCLUDED.contact_title,
    address = EXCLUDED.address,
    city = EXCLUDED.city,
    region = EXCLUDED.region,
    postal_code = EXCLUDED.postal_code,
    country = EXCLUDED.country,
    phone = EXCLUDED.phone;

INSERT INTO products (id, category_id, supplier_id, name, quantity_per_unit, unit_price, units_in_stock, units_on_order, reorder_level, discontinued) VALUES
    (1, 1, 1, 'Chai', '10 boxes x 20 bags', 18.00, 39, 0, 10, FALSE),
    (2, 1, 1, 'Chang', '24 - 12 oz bottles', 19.00, 17, 40, 25, FALSE),
    (3, 2, 2, 'Aniseed Syrup', '12 - 550 ml bottles', 10.00, 13, 70, 25, FALSE),
    (4, 3, 3, 'Chef Anton''s Cajun Seasoning', '48 - 6 oz jars', 22.00, 53, 0, 0, FALSE),
    (5, 3, 3, 'Grandma''s Boysenberry Spread', '12 - 8 oz jars', 25.00, 120, 0, 25, FALSE)
ON CONFLICT (id) DO UPDATE SET
    category_id = EXCLUDED.category_id,
    supplier_id = EXCLUDED.supplier_id,
    name = EXCLUDED.name,
    quantity_per_unit = EXCLUDED.quantity_per_unit,
    unit_price = EXCLUDED.unit_price,
    units_in_stock = EXCLUDED.units_in_stock,
    units_on_order = EXCLUDED.units_on_order,
    reorder_level = EXCLUDED.reorder_level,
    discontinued = EXCLUDED.discontinued;

SELECT setval(pg_get_serial_sequence('categories', 'id'), COALESCE((SELECT MAX(id) FROM categories), 1));
SELECT setval(pg_get_serial_sequence('suppliers', 'id'), COALESCE((SELECT MAX(id) FROM suppliers), 1));
SELECT setval(pg_get_serial_sequence('products', 'id'), COALESCE((SELECT MAX(id) FROM products), 1));
