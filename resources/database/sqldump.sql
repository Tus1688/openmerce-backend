--  Copyright (c) 2023. Tus1688
--
--  Permission is hereby granted, free of charge, to any person obtaining a copy
--  of this software and associated documentation files (the "Software"), to deal
--  in the Software without restriction, including without limitation the rights
--  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
--  copies of the Software, and to permit persons to whom the Software is
--  furnished to do so, subject to the following conditions:
--
--  The above copyright notice and this permission notice shall be included in all
--  copies or substantial portions of the Software.
--
--  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
--  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
--  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
--  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
--  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
--  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
--  SOFTWARE.

CREATE TABLE customers(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    email VARCHAR(72) UNIQUE NOT NULL,
    hashed_password BINARY(60) NOT NULL,
    phone_number VARCHAR(15) UNIQUE,
    first_name VARCHAR(50) NOT NULL,
    last_name VARCHAR(50) NOT NULL,
    birth_date DATETIME,
    gender VARCHAR(6),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE auth_logs(
      id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
      customer_refer BINARY(16) NOT NULL,
      timestamp datetime,
      jti varchar(50),
      user_agent varchar(255),
      ip_address varchar(15),
      action varchar(8),
      INDEX customer_refer_idx(customer_refer),
      INDEX timestamp_idx(timestamp),
      FOREIGN KEY (customer_refer) REFERENCES customers(id)
);

CREATE TABLE areas (
    code varchar(13) PRIMARY KEY,
    name varchar(100) NOT NULL ,
    FULLTEXT INDEX areas_name_idx(name),
    INDEX areas_code_idx(code)
);

CREATE TABLE customer_addresses(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    customer_refer BINARY(16) NOT NULL,
    label VARCHAR(30) NOT NULL,
    full_address VARCHAR(255) NOT NULL,
    note VARCHAR(45),
    recipient_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    shipping_area_refer MEDIUMINT UNSIGNED NOT NULL,
    postal_code varchar(5) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    -- hard delete for now
    INDEX customer_refer_idx(customer_refer),
    UNIQUE (customer_refer, label),
    FOREIGN KEY (customer_refer) REFERENCES customers(id),
    FOREIGN KEY (shipping_area_refer) REFERENCES shipping_areas(id)
);

CREATE TABLE categories(
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) UNIQUE NOT NULL,
    description VARCHAR(255) NOT NULL,
    homepage_visibility BOOLEAN DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    INDEX category_name_idx(name),
    INDEX category_homepage_visibility_idx(homepage_visibility)
);

CREATE TABLE products(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    name VARCHAR(85) UNIQUE NOT NULL,
    description VARCHAR(300) NOT NULL,
    price INT UNSIGNED NOT NULL,
    weight DECIMAL(10,2) NOT NULL,
    category_refer INT UNSIGNED NOT NULL,
    cumulative_review DECIMAL(2,1) DEFAULT 0,
    length SMALLINT UNSIGNED NOT NULL,
    width SMALLINT UNSIGNED NOT NULL,
    height SMALLINT UNSIGNED NOT NULL,
    created_at datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime,
    deleted_at datetime,
    INDEX product_check_exist_idx(id, deleted_at),
    INDEX product_category_idx(category_refer, deleted_at),
    FULLTEXT INDEX product_name_idx(name),
    FOREIGN KEY (category_refer) REFERENCES categories(id)
);

CREATE TABLE product_images(
    id BINARY(16)  PRIMARY KEY,
    product_refer BINARY(16) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX product_images_product_refer_idx(product_refer),
    INDEX product_images_created_at_idx(created_at),
    FOREIGN KEY (product_refer) REFERENCES products(id)
);

CREATE TABLE inventories(
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    product_refer BINARY(16) NOT NULL,
    quantity INT UNSIGNED NOT NULL,
    updated_at DATETIME,
    INDEX inventories_product_refer_idx(product_refer),
    FOREIGN KEY (product_refer) REFERENCES products(id)
);

CREATE TABLE cart_items(
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    product_refer BINARY(16) NOT NULL,
    customer_refer BINARY(16) NOT NULL,
    quantity SMALLINT UNSIGNED NOT NULL,
    checked BOOLEAN DEFAULT FALSE,
    INDEX cart_items_customer_idx(customer_refer),
    UNIQUE(product_refer, customer_refer),
    FOREIGN KEY (product_refer) REFERENCES products(id),
    FOREIGN KEY (customer_refer) REFERENCES customers(id)
);

CREATE TABLE wishlists(
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    product_refer BINARY(16) NOT NULL,
    customer_refer BINARY(16) NOT NULL,
    INDEX wishlists_customer_idx(customer_refer),
    UNIQUE(product_refer, customer_refer),
    FOREIGN KEY (product_refer) REFERENCES products(id),
    FOREIGN KEY (customer_refer) REFERENCES customers(id)
);

CREATE TABLE orders(
    id                     BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    customer_refer         BINARY(16)   NOT NULL,
    customer_address_refer BINARY(16)   NOT NULL,
    courier_code           VARCHAR(255) NOT NULL,
    courier_tracking_code  VARCHAR(255) NULL,
    freight_cost           INT UNSIGNED NOT NULL,
    item_cost              INT UNSIGNED NOT NULL,
    gross_amount           INT UNSIGNED NOT NULL,
    # transaction_status can be capture, settlement, pending, deny, cancel, expire, refund, partial_refund, authorize
    transaction_status     VARCHAR(255) NULL,
    # status_description show the reason of the transaction_status
    status_description     VARCHAR(255) NULL,
    # payment_token is the token that will be retrieved from the payment gateway
    payment_token          VARCHAR(255) NULL,
    # payment_redirect_url is the url that will be redirected to the payment gateway
    payment_redirect_url   VARCHAR(255) NULL,
    # is_paid is the flag to indicate whether the order is paid or not
    is_paid                BOOLEAN  DEFAULT FALSE,
    # is_shipped is the flag to indicate whether the order is shipped or not
    is_shipped             BOOLEAN  DEFAULT FALSE,
    # is_cancelled is the flag to indicate whether the order is cancelled or not
    is_cancelled           BOOLEAN  DEFAULT FALSE,
    # need_refund is the flag to indicate whether the order is need to be refunded or not (if the quantity is not enough)
    need_refund            BOOLEAN  DEFAULT FALSE,
    payment_type           VARCHAR(255) NULL,
    created_at             datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at             datetime,
    deleted_at             datetime,
    INDEX awaiting_orders_customer_refer_idx (customer_refer),
    INDEX is_paid_idx (is_paid, customer_refer),
    INDEX is_shipped_idx (is_shipped),
    FOREIGN KEY (customer_refer) REFERENCES customers (id),
    FOREIGN KEY (customer_address_refer) REFERENCES customer_addresses (id)
);

CREATE TABLE order_items(
    id                 BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_refer        BIGINT UNSIGNED   NOT NULL,
    product_refer      BINARY(16)        NOT NULL,
    on_buy_name        VARCHAR(85)       NOT NULL,
    on_buy_description VARCHAR(300)      NOT NULL,
    on_buy_price       INT UNSIGNED      NOT NULL,
    on_buy_weight      DECIMAL(10,2) NOT NULL,
    quantity           SMALLINT UNSIGNED NOT NULL,
    INDEX order_details_order_refer (order_refer),
    FOREIGN KEY (order_refer) REFERENCES orders(id),
    FOREIGN KEY (product_refer) REFERENCES products (id)
);

CREATE TABLE reviews(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    order_item_refer BIGINT UNSIGNED NOT NULL,
    product_refer BINARY(16) NOT NULL,
    rating TINYINT UNSIGNED NOT NULL,
    review VARCHAR(255),
    INDEX reviews_product_refer_idx (product_refer),
    UNIQUE (order_item_refer),
    FOREIGN KEY (order_item_refer) REFERENCES order_items(id),
    FOREIGN KEY (product_refer) REFERENCES products(id)
);

CREATE TABLE staffs(
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(32) UNIQUE NOT NULL,
    hashed_password BINARY(60) NOT NULL,
    name VARCHAR(100) NOT NULL,
    fin_user BOOLEAN NOT NULL DEFAULT FALSE,
    inv_user BOOLEAN NOT NULL DEFAULT FALSE,
    sys_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE homepage_banner(
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    file_name varchar(41) NOT NULL,
    href varchar(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME
);

CREATE TABLE logs(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    log_level VARCHAR(7) NOT NULL,
    info VARCHAR(255) NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE blacklist_domains (
    domain_name VARCHAR(255) UNIQUE
)
PARTITION BY RANGE COLUMNS(domain_name) (
    PARTITION p0 VALUES LESS THAN ('1'),
    PARTITION p1 VALUES LESS THAN ('2'),
    PARTITION p2 VALUES LESS THAN ('3'),
    PARTITION p3 VALUES LESS THAN ('4'),
    PARTITION p4 VALUES LESS THAN ('5'),
    PARTITION p5 VALUES LESS THAN ('6'),
    PARTITION p6 VALUES LESS THAN ('7'),
    PARTITION p7 VALUES LESS THAN ('8'),
    PARTITION p8 VALUES LESS THAN ('9'),
    PARTITION p9 VALUES LESS THAN ('a'),
    PARTITION p10 VALUES LESS THAN ('b'),
    PARTITION p11 VALUES LESS THAN ('c'),
    PARTITION p12 VALUES LESS THAN ('d'),
    PARTITION p13 VALUES LESS THAN ('e'),
    PARTITION p14 VALUES LESS THAN ('f'),
    PARTITION p15 VALUES LESS THAN ('g'),
    PARTITION p16 VALUES LESS THAN ('h'),
    PARTITION p17 VALUES LESS THAN ('i'),
    PARTITION p18 VALUES LESS THAN ('j'),
    PARTITION p19 VALUES LESS THAN ('k'),
    PARTITION p20 VALUES LESS THAN ('l'),
    PARTITION p21 VALUES LESS THAN ('m'),
    PARTITION p22 VALUES LESS THAN ('n'),
    PARTITION p23 VALUES LESS THAN ('o'),
    PARTITION p24 VALUES LESS THAN ('p'),
    PARTITION p25 VALUES LESS THAN ('q'),
    PARTITION p26 VALUES LESS THAN ('r'),
    PARTITION p27 VALUES LESS THAN ('s'),
    PARTITION p28 VALUES LESS THAN ('t'),
    PARTITION p29 VALUES LESS THAN ('u'),
    PARTITION p30 VALUES LESS THAN ('v'),
    PARTITION p31 VALUES LESS THAN ('w'),
    PARTITION p32 VALUES LESS THAN ('x'),
    PARTITION p33 VALUES LESS THAN ('y'),
    PARTITION p34 VALUES LESS THAN ('z'),
    PARTITION p35 VALUES LESS THAN (MAXVALUE)
);