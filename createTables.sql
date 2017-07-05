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

-- user=bp pass=monitor
-- test data
INSERT INTO users (username, password) VALUES ('bp', '$2a$10$OEYDcQEQqikOqaF0k5PLBu8gq0rJQ9RARxe8eLX1reyL15Qf4gaya');
INSERT INTO measurements (user_id, systolic, diastolic, pulse, created_at)
VALUES (1, 127, 79, 68, '2017-06-02 20:10:45'), (1, 125, 89, 86, '2017-06-02 22:40:30'),
  (1, 123, 86, 75, '2017-06-03 00:51:12'), (1, 138, 96, 73, '2017-06-03 04:40:36'),
  (1, 128, 88, 102, '2017-06-14 00:15:27'), (1, 125, 85, 74, '2017-06-20 01:36:27'),
  (1, 119, 80, 81, '2017-06-22 07:57:26'), (1, 126, 86, 67, '2017-06-28 14:41:07'),
  (1, 121, 83, 71, '2017-07-01 16:27:49'), (1, 131, 94, 73, '2017-07-04 19:49:07');
