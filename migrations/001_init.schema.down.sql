-- Drop foreign-key dependent tables first
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS landlords;
DROP TABLE IF EXISTS agencies;
DROP TABLE IF EXISTS otp_codes;

-- Drop independent tables
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS clients;