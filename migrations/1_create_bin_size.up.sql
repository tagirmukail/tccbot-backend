CREATE TABLE IF NOT EXISTS bin_size (
    bin VARCHAR(80) PRIMARY KEY NOT NULL
);

INSERT INTO bin_size(bin) VALUES ('1m'),( '5m'),( '1h'),('1d');
