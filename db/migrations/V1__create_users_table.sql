CREATE TABLE users (
                       id INT PRIMARY KEY,
                       name VARCHAR(255) NOT NULL,
                       status VARCHAR(50),
                       created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);