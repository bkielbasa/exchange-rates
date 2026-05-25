CREATE TABLE rates (
    id          BIGINT UNSIGNED  NOT NULL AUTO_INCREMENT,
    currency    CHAR(3)          NOT NULL,
    rate        DECIMAL(20,8)    NOT NULL,
    fetched_at  DATETIME(6)      NOT NULL,
    PRIMARY KEY (id),
    KEY idx_currency_fetched_at (currency, fetched_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
