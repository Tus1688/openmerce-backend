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
    name varchar(100),
    FULLTEXT INDEX areas_name_idx(name)
);

CREATE TABLE customer_addresses(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    customer_refer BINARY(16) NOT NULL,
    label VARCHAR(30) NOT NULL,
    full_address VARCHAR(255) NOT NULL,
    note VARCHAR(45),
    recipient_name VARCHAR(50) NOT NULL,
    phone_number VARCHAR(15) NOT NULL,
    subdistrict varchar(13) NOT NULL,
    district varchar(13) NOT NULL,
    city varchar(13) NOT NULL,
    province varchar(13) NOT NULL,
    postal_code varchar(5) NOT NULL,
    INDEX customer_refer_idx(customer_refer),
    FOREIGN KEY (customer_refer) REFERENCES customers(id),
    FOREIGN KEY (subdistrict) REFERENCES areas(code),
    FOREIGN KEY (district) REFERENCES areas(code),
    FOREIGN KEY (city) REFERENCES areas(code),
    FOREIGN KEY (province) REFERENCES areas(code)
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
    FOREIGN KEY (product_refer) REFERENCES products(id),
    FOREIGN KEY (customer_refer) REFERENCES customers(id)
);

CREATE TABLE couriers(
    id INT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(25) UNIQUE NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    taxid VARCHAR(16) NOT NULL,
    email VARCHAR(72) NOT NULL,
    phone_number VARCHAR(16) NOT NULL,
    rep_name VARCHAR(60) NOT NULL
);

CREATE TABLE orders(
    -- no auto increment to maintain data integrity
    id BIGINT UNSIGNED PRIMARY KEY,
    customer_refer BINARY(16) NOT NULL,
    customer_address_refer BINARY(16) NOT NULL,
    courier_refer INT UNSIGNED NOT NULL,
    total_weight INT UNSIGNED NOT NULL,
    freight_cost INT UNSIGNED,
    item_cost INT UNSIGNED NOT NULL,
    gross_amount INT UNSIGNED,
    freight_booking_id VARCHAR(25),
    created_at datetime DEFAULT CURRENT_TIMESTAMP,
    updated_at datetime,
    deleted_at datetime,
    INDEX orders_customer_refer_idx(customer_refer),
    FOREIGN KEY (customer_refer) REFERENCES customers(id),
    FOREIGN KEY (customer_address_refer) REFERENCES customer_addresses(id),
    FOREIGN KEY (courier_refer) REFERENCES couriers(id)
);

CREATE TABLE payments(
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_refer BIGINT UNSIGNED NOT NULL,
    payment_type VARCHAR(100) NOT NULL,
    transaction_id VARCHAR(100),
    gross_amount INT UNSIGNED,
    status VARCHAR(20),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    INDEX payments_order_refer_idx(order_refer),
    FOREIGN KEY (order_refer) REFERENCES orders(id)
);

CREATE TABLE shipping_logs(
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    waybill VARCHAR(25) NOT NULL,
    booking_id VARCHAR(25) NOT NULL,
    message VARCHAR(500) NOT NULL,
    tracking_code VARCHAR(5) NOT NULL,
    timestamp DATETIME,
    INDEX shipping_logs_booking_id_idx (booking_id)
    -- FOREIGN KEY (booking_id) REFERENCES orders(freight_booking_id)
);

CREATE TABLE order_details(
    id BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
    order_refer BIGINT UNSIGNED NOT NULL,
    product_refer BINARY(16) NOT NULL,
    on_buy_name VARCHAR(85) NOT NULL,
    on_buy_description VARCHAR(300) NOT NULL,
    on_buy_price INT UNSIGNED NOT NULL,
    on_buy_weight SMALLINT UNSIGNED NOT NULL,
    quantity SMALLINT UNSIGNED NOT NULL,
    INDEX order_details_order_refer(order_refer),
    FOREIGN KEY (order_refer) REFERENCES orders(id),
    FOREIGN KEY (product_refer) REFERENCES products(id)
);

CREATE TABLE reviews(
    id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
    order_detail_refer BIGINT UNSIGNED NOT NULL,
    product_refer BINARY(16) NOT NULL,
    rating TINYINT UNSIGNED NOT NULL,
    review VARCHAR(255),
    INDEX reviews_product_refer_idx (product_refer),
    UNIQUE (order_detail_refer, product_refer),
    FOREIGN KEY (order_detail_refer) REFERENCES order_details(id),
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