CREATE USER docker;
CREATE DATABASE docker;
GRANT ALL PRIVILEGES ON DATABASE docker to docker;

CREATE TABLE products
(
	ID SERIAL,
	NAME TEXT NOT NULL
);

INSERT INTO products (NAME) VALUES('Hose');
INSERT INTO products (NAME) VALUES('Schuhe');
