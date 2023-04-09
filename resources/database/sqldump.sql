CREATE TABLE
    customers(
        id BINARY(16) PRIMARY KEY DEFAULT (UUID_TO_BIN(UUID())),
        email VARCHAR(72) UNIQUE NOT NULL,
        hashed_password BINARY(60) NOT NULL,
        phone_number VARCHAR(15) UNIQUE,
        first_name VARCHAR(50) NOT NULL,
        last_name VARCHAR(50) NOT NULL,
        birth_date DATETIME,
        gender VARCHAR(6)
    );

