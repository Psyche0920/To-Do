CREATE TABLE tasks (
    -- Generate a unique task ID; gaps after failed inserts are normal.
    id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    -- Reject NULL, more than 200 characters, and empty or space-only titles.
    -- btrim checks a trimmed value; it does not change the stored title.
    title VARCHAR(200) NOT NULL CHECK (length(btrim(title)) > 0),
    -- Use false when an inserted task omits its completion status.
    completed BOOLEAN NOT NULL DEFAULT FALSE
);
