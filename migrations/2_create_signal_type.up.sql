CREATE TABLE IF NOT EXISTS signal_type (signal VARCHAR(100) PRIMARY KEY NOT NULL);

INSERT INTO signal_type(signal) VALUES ('SMA'),('EMA'),('WMA'),('RSI'),('BolingerBand'),('MACD');
