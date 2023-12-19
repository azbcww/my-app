CREATE DATABASE IF NOT EXISTS events;
USE events;
CREATE TABLE content (
    userName varchar(255) NOT NULL,
    startDate VARCHAR(255) NOT NULL,
    endDate VARCHAR(255) NOT NULL,
    title TEXT NOT NULL
);
ALTER TABLE content ADD INDEX idx (userNAME);