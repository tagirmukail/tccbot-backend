CREATE TABLE signals(
                        id         integer PRIMARY KEY AUTOINCREMENT,
                        bin        VARCHAR(100) NOT NULL DEFAULT ('') REFERENCES bin_size(bin),
                        signal_t   VARCHAR(100) NOT NULL DEFAULT ('') REFERENCES signal_type(signal),
                        n          int,
                        macd_fast  int,
                        macd_slow  int,
                        macd_sig   int,
                        timestamp  timestamp   NOT NULL,
                        signal_v   float4,
                        bbtl 	   float4,
                        bbml	   float4,
                        bbbl 	   float4,
                        macd_v     float4,
                        macd_h_v   float4,
                        created_at bigint NOT NULL,
                        updated_at bigint NOT NULL
);
