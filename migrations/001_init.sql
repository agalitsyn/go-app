CREATE TABLE article (
    id          SERIAL      PRIMARY KEY,
    title       text        NOT NULL,
    slug        text        UNIQUE NOT NULL
);
