CREATE DATABASE IF NOT EXISTS rate_app;

USE rate_app;

CREATE TABLE IF NOT EXISTS rates (
  date datetime,
  currency varchar(10),
  rate double
);
