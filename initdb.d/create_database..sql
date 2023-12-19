CREATE DATABASE IF NOT EXISTS events;
USE events;
CREATE TABLE content (
    Username varchar(255) NOT NULL,
    StartDate VARCHAR(255) NOT NULL,
    EndDate VARCHAR(255) NOT NULL,
    Title TEXT NOT NULL
);
ALTER TABLE content ADD INDEX idx (userNAME);