
-- Initialize User Table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    created_at BIGINT DEFAULT (EXTRACT(EPOCH FROM CURRENT_TIMESTAMP))
);

-- Initialize Orders Table
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    symbol VARCHAR(20) NOT NULL,
    quantity FLOAT NOT NULL,
    price FLOAT NOT NULL,
    stop_loss BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'open',
    created_at BIGINT DEFAULT (EXTRACT(EPOCH FROM CURRENT_TIMESTAMP))
);

-- Initialize Stop Loss Criteria Table
CREATE TABLE IF NOT EXISTS stop_loss_criteria (
    id SERIAL PRIMARY KEY,
    order_id INT REFERENCES orders(id),
    stop_loss_price FLOAT NOT NULL,
    expiry_date BIGINT NOT NULL
);