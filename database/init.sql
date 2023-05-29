CREATE DATABASE rate_app;

CREATE USER 'root'@'localhost' IDENTIFIED BY 'hrk_database';
GRANT ALL PRIVILEGES ON rate_app.* TO 'root'@'localhost';
FLUSH PRIVILEGES;

USE rate_app;

CREATE TABLE IF NOT EXISTS rates (
    date datetime,
    currency varchar(10),
    rate double
);
