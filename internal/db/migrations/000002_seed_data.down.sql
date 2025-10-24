DELETE FROM products WHERE id BETWEEN 1 AND 5;
DELETE FROM suppliers WHERE id BETWEEN 1 AND 3;
DELETE FROM categories WHERE id BETWEEN 1 AND 3;

SELECT setval(pg_get_serial_sequence('products', 'id'), COALESCE((SELECT MAX(id) FROM products), 0));
SELECT setval(pg_get_serial_sequence('suppliers', 'id'), COALESCE((SELECT MAX(id) FROM suppliers), 0));
SELECT setval(pg_get_serial_sequence('categories', 'id'), COALESCE((SELECT MAX(id) FROM categories), 0));
