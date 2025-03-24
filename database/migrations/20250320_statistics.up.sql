CREATE TABLE statistics
(
    name TEXT NOT NULL,
    value INTEGER NOT NULL,
    percentage NUMERIC(5, 2),
    updated_at TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (name)
);
