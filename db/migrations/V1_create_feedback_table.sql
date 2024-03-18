CREATE TABLE feedback (
                         id SERIAL PRIMARY KEY,
                         message TEXT NOT NULL,
                         user_id INTEGER NOT NULL,
                         FOREIGN KEY (user_id) REFERENCES users(id)
);