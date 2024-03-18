CREATE TABLE money_events (
                              id SERIAL PRIMARY KEY,
                              amount FLOAT NOT NULL,
                              currency VARCHAR(3) NOT NULL,
                              comment TEXT,
                              tag VARCHAR(50),
                              created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              user_id INTEGER NOT NULL,
                              FOREIGN KEY (user_id) REFERENCES users(id)
);
