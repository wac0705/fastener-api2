-- db/migrations/000001_initial_schema.down.sql

-- 刪除 product_definitions 表
DROP TABLE IF EXISTS product_definitions CASCADE;

-- 刪除 product_categories 表
DROP TABLE IF EXISTS product_categories CASCADE;

-- 刪除 role_menus 表
DROP TABLE IF EXISTS role_menus CASCADE;

-- 刪除 menus 表
DROP TABLE IF EXISTS menus CASCADE;

-- 刪除 customers 表
DROP TABLE IF EXISTS customers CASCADE;

-- 刪除 companies 表
DROP TABLE IF EXISTS companies CASCADE;

-- 刪除 accounts 表
DROP TABLE IF EXISTS accounts CASCADE;

-- 刪除 role_permissions 表
DROP TABLE IF EXISTS role_permissions CASCADE;

-- 刪除 permissions 表
DROP TABLE IF EXISTS permissions CASCADE;

-- 刪除 roles 表
DROP TABLE IF EXISTS roles CASCADE;
