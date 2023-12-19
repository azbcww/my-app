CREATE DATABASE IF NOT EXISTS events;
USE events;
CREATE TABLE content (
    id BIGINT NOT NULL PRIMARY KEY,
    userID BIGINT NOT NULL,
    startDate VARCHAR(255) NOT NULL,
    endDate VARCHAR(255) NOT NULL,
    title TEXT
);
ALTER TABLE content ADD INDEX idx (userID);