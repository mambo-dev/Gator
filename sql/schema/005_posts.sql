-- +goose Up
CREATE TABLE posts (
    id uuid PRIMARY KEY,                     
    created_at TIMESTAMP NOT NULL DEFAULT NOW(), 
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(), 
    title TEXT NOT NULL,                      
    url TEXT NOT NULL UNIQUE,                  
    description TEXT,                           
    published_at TIMESTAMP,                
    feed_id uuid NOT NULL,                   
    FOREIGN KEY (feed_id) REFERENCES feeds (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;