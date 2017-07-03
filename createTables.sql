-- DELETE FROM mysql.user WHERE user='bpuser'; FLUSH PRIVILEGES;
CREATE USER 'bpuser'@'localhost'
  IDENTIFIED BY 'bppass';
FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'bpuser'@'localhost';
FLUSH PRIVILEGES;
CREATE USER 'bpuser'@'%'
  IDENTIFIED BY 'bppass';
FLUSH PRIVILEGES;
GRANT ALL PRIVILEGES ON *.* TO 'bpuser'@'%';
FLUSH PRIVILEGES;

CREATE DATABASE bpmonitor;
USE bpmonitor;

DROP TABLE IF EXISTS measurements;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
  id         BIGINT UNIQUE       NOT NULL AUTO_INCREMENT,
  username   VARCHAR(255) UNIQUE NOT NULL,
  password   VARCHAR(255)        NOT NULL DEFAULT "",
  apikey     VARCHAR(255) UNIQUE,
  session_id VARCHAR(255) UNIQUE,
  PRIMARY KEY (id)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

CREATE TABLE measurements (
  id         BIGINT        NOT NULL AUTO_INCREMENT,
  user_id    BIGINT        NOT NULL,
  systolic   INT,
  diastolic  INT,
  pulse      INT,
  notes      VARCHAR(1024) NOT NULL DEFAULT "",
  created_at TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE KEY (user_id, created_at),
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES users (id)
)
  ENGINE = InnoDB
  DEFAULT CHARSET = utf8;

INSERT INTO users (username, password) VALUES ('bp', 'bp');
