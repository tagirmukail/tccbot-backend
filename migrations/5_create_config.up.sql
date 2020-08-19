CREATE TABLE IF NOT EXISTS global_config (
    id         integer PRIMARY KEY AUTOINCREMENT,
    db_path    varchar(300),
    order_proc_period_in_sec integer,
    created_at timestamp NOT NULL,
    updated_at timestamp NOT NULL,
    admin_id INTEGER,
    FOREIGN KEY(admin_id) REFERENCES admin(id)
);

CREATE TABLE IF NOT EXISTS exchanges_api_settings (
    id         integer PRIMARY KEY AUTOINCREMENT,
    exchange   varchar(90) unique not null,
    enable     boolean not null,
    test       boolean not null,
    ping_sec   integer not null,
    timeout_sec integer not null,
    retry_sec   integer not null,
    buffer_size integer not null,
    currency   varchar(100) not null,
    symbol     varchar(100) not null,
    order_type varchar(100) not null,
    max_amount float4 not null,
    limit_contracts_cnt integer not null,
    sell_order_coef float4 not null,
    buy_order_coef float4 not null,
    global_id INTEGER,
    FOREIGN KEY(global_id) REFERENCES global_config(id)
);

CREATE TABLE IF NOT EXISTS exchange_access (
    id         integer PRIMARY KEY AUTOINCREMENT,
    exchange   varchar(90) not null,
    test       boolean not null,
    key        varchar(150) not null,
    secret     varchar(300) not null,
    global_id INTEGER,
    FOREIGN KEY(global_id) REFERENCES global_config(id)
);

CREATE TABLE IF NOT EXISTS admin (
    id         integer PRIMARY KEY AUTOINCREMENT,
    exchange   varchar(90) not null,
    username   varchar(90) not null,
    secret_token varchar(290) not null,
    token varchar(290) not null,
    global_id INTEGER,
    FOREIGN KEY(global_id) REFERENCES global_config(id)
);

CREATE TABLE IF NOT EXISTS scheduler (
    id         integer PRIMARY KEY AUTOINCREMENT,
    type       varchar(90) unique not null,
    enable     boolean not null,
    price_trailing float4 not null,
    profit_close_btc float4 not null,
    loss_close_btc float4 not null,
    profit_pnl_diff float4 not null,
    loss_pnl_diff float4 not null,
    global_id INTEGER,
    FOREIGN KEY(global_id) REFERENCES global_config(id)
);

CREATE TABLE IF NOT EXISTS strategies_config (
    id         integer PRIMARY KEY AUTOINCREMENT,
    bin        VARCHAR(100) NOT NULL DEFAULT ('') REFERENCES bin_size(bin),
    enable_rsi_bb boolean not null,
    retry_process_count integer not null,
    get_candles_count integer not null,

    trend_filter_enable boolean not null,
    candles_filter_enable boolean not null,
    max_filter_trend_count integer not null,
    max_candles_filter_count integer not null,

    bb_last_candles_count integer not null,

    rsi_count integer not null,
    rsi_min_border integer not null,
    rsi_max_border integer not null,
    rsi_trade_coef float4 not null,

    macd_fast_count integer not null,
    macd_slow_count integer not null,
    macd_sig_count integer not null,

    global_id INTEGER,
    FOREIGN KEY(global_id) REFERENCES global_config(id)
)
