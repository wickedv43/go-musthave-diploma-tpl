-- queries/queries.sql

-- name: CreateUser :one
INSERT INTO users (login, password, balance_current, balance_withdrawn)
VALUES ($1, $2, $3, $4)
    RETURNING id, login, password, balance_current, balance_withdrawn;

-- name: GetUserByID :one
SELECT id, login, password, balance_current, balance_withdrawn
FROM users
WHERE id = $1;

-- name: GetUserByLogin :one
SELECT id, login, password, balance_current, balance_withdrawn
FROM users
WHERE login = $1;

-- name: CreateOrder :one
INSERT INTO orders (number, user_id, status, accrual, uploaded_at)
VALUES ($1, $2, $3, $4, $5)
    RETURNING number, user_id, status, accrual, uploaded_at;

-- name: GetOrdersByUserID :many
SELECT number, user_id, status, accrual, uploaded_at
FROM orders
WHERE user_id = $1;

-- name: CreateBill :one
INSERT INTO bills (order_number, user_id, sum, processed_at)
VALUES ($1, $2, $3, $4)
    RETURNING id, order_number, user_id, sum, processed_at;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2
WHERE number = $1;

-- name: GetOrderByNumber :one
SELECT number, user_id, status, accrual, uploaded_at
FROM orders
WHERE number = $1;

-- name: GetBillsByUserID :many
SELECT id, order_number, user_id, sum, processed_at
FROM bills
WHERE user_id = $1
ORDER BY processed_at DESC;

-- name: GetAllBills :many
SELECT id, order_number, user_id, sum, processed_at
FROM bills
ORDER BY processed_at DESC;


-- name: GetBillByID :one
SELECT id, order_number, user_id, sum, processed_at
FROM bills
WHERE id = $1;

-- name: UpdateUserBalance :exec
UPDATE users
SET balance_current = $2,
    balance_withdrawn = $3
WHERE id = $1;

