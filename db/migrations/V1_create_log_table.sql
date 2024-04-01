CREATE TABLE usage_log (
   id SERIAL PRIMARY KEY,
   user_id int NOT NULL,
   reply_type varchar(20),
   created TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);