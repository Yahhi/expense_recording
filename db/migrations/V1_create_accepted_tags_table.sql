CREATE TABLE accepted_tags (
                         id SERIAL PRIMARY KEY,
                         tag VARCHAR(50) NOT NULL,
                         user_id INTEGER NOT NULL,
                         FOREIGN KEY (user_id) REFERENCES users(id)
);