CREATE TABLE targets (
                         id SERIAL PRIMARY KEY,
                         tag VARCHAR(50) NOT NULL,
                         amount FLOAT NOT NULL,
                         period_start TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                         period_end TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                         user_id INTEGER NOT NULL,
                         FOREIGN KEY (user_id) REFERENCES users(id)
);
