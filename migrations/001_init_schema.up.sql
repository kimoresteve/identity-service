-- Clients table (base for all identities)
CREATE TABLE IF NOT EXISTS clients
(
    id          INT AUTO_INCREMENT PRIMARY KEY,
    uuid        VARCHAR(36) NOT NULL UNIQUE,
    type        ENUM('agency', 'landlord', 'user') NOT NULL,
    email       VARCHAR(255) UNIQUE,
    contact     VARCHAR(50) UNIQUE,
    password    VARCHAR(255),
    is_verified BOOLEAN   DEFAULT FALSE,
    is_active   BOOLEAN   DEFAULT TRUE,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Agencies table (extends clients)
CREATE TABLE  IF NOT EXISTS agencies
(
    id       INT PRIMARY KEY,
    name     VARCHAR(255) NOT NULL,
    address  TEXT,
    tax_id   VARCHAR(100),
    logo_url VARCHAR(255),
    FOREIGN KEY (id) REFERENCES clients (id) ON DELETE CASCADE
);

-- Landlords table (extends clients)
CREATE TABLE  IF NOT EXISTS landlords
(
    id        INT PRIMARY KEY,
    name      VARCHAR(255) NOT NULL,
    address   TEXT,
    agency_id INT NULL, -- NULL if independent landlord
    FOREIGN KEY (id) REFERENCES clients (id) ON DELETE CASCADE,
    FOREIGN KEY (agency_id) REFERENCES agencies (id) ON DELETE SET NULL
);

-- Users table (belongs to either agency or landlord)
CREATE TABLE  IF NOT EXISTS users
(
    id         INT PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name  VARCHAR(100) NOT NULL,
    position   VARCHAR(100),
    owner_id   INT          NOT NULL, -- agency_id or landlord_id
    owner_type ENUM('agency', 'landlord') NOT NULL,
    FOREIGN KEY (id) REFERENCES clients (id) ON DELETE CASCADE
);

-- OTP codes (for all client types)
CREATE TABLE   IF NOT EXISTS otp_codes
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    client_id  INT         NOT NULL,
    otp        VARCHAR(10) NOT NULL,
    created_at  DATETIME   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    purpose    ENUM('activation', 'reset', '2fa')  NOT NULL DEFAULT 'activation',
    expires_at TIMESTAMP   NOT NULL,
    used       BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (client_id) REFERENCES clients (id) ON DELETE CASCADE
);

-- Permissions (for future implementation)
CREATE TABLE  IF NOT EXISTS permissions
(
    id          INT AUTO_INCREMENT PRIMARY KEY,
    name        VARCHAR(100) NOT NULL UNIQUE,
    description TEXT
);

-- Roles (for future implementation)
CREATE TABLE  IF NOT EXISTS roles
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(100) NOT NULL,
    owner_id   INT NULL, -- NULL for system-wide roles
    owner_type ENUM('system', 'agency', 'landlord') NOT NULL DEFAULT 'system'
);

-- Role-Permission mapping (for future implementation)
CREATE TABLE  IF NOT EXISTS  role_permissions
(
    role_id       INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE
);

-- User-Role mapping (for future implementation)
CREATE TABLE  IF NOT EXISTS  user_roles
(
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE
);