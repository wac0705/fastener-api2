-- db/migrations/000001_initial_schema.up.sql

-- 建立 roles 表
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL, -- 例如: 'admin', 'finance', 'customer'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 建立 permissions 表
CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL, -- 例如: 'company:read', 'account:create'
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 建立 role_permissions 表 (多對多關係)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- 建立 accounts 表
CREATE TABLE IF NOT EXISTS accounts (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role_id INT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT -- 避免刪除角色時級聯刪除用戶
);

-- 建立 companies 表
CREATE TABLE IF NOT EXISTS companies (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 建立 customers 表
CREATE TABLE IF NOT EXISTS customers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    contact_person VARCHAR(255),
    email VARCHAR(255),
    phone VARCHAR(50),
    company_id INT, -- 關聯到公司
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (company_id) REFERENCES companies(id) ON DELETE SET NULL
);

-- 建立 menus 表 (系統選單)
CREATE TABLE IF NOT EXISTS menus (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    path VARCHAR(255) UNIQUE NOT NULL, -- 前端路由路徑
    icon VARCHAR(50), -- 選單圖標
    parent_id INT, -- 父選單 ID，用於嵌套選單
    display_order INT NOT NULL DEFAULT 0, -- 顯示順序
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (parent_id) REFERENCES menus(id) ON DELETE SET NULL
);

-- 建立 role_menus 表 (角色與選單關係)
CREATE TABLE IF NOT EXISTS role_menus (
    role_id INT NOT NULL,
    menu_id INT NOT NULL,
    PRIMARY KEY (role_id, menu_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (menu_id) REFERENCES menus(id) ON DELETE CASCADE
);

-- 建立 product_categories 表
CREATE TABLE IF NOT EXISTS product_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 建立 product_definitions 表
CREATE TABLE IF NOT EXISTS product_definitions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category_id INT NOT NULL,
    unit VARCHAR(50),
    price DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (category_id) REFERENCES product_categories(id) ON DELETE RESTRICT
);

-- 插入初始數據 (這些數據將在應用程式啟動時由 resetadmin 工具或手動插入)
-- 初始角色
INSERT INTO roles (name) VALUES ('admin') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('finance') ON CONFLICT (name) DO NOTHING;
INSERT INTO roles (name) VALUES ('user') ON CONFLICT (name) DO NOTHING;

-- 初始權限 (需要根據你的 API 路由和業務邏輯詳細定義)
-- 帳戶管理
INSERT INTO permissions (name, description) VALUES ('account:read', 'Allow reading account information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('account:create', 'Allow creating new accounts') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('account:update', 'Allow updating account information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('account:delete', 'Allow deleting accounts') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('account:update_password', 'Allow updating account password') ON CONFLICT (name) DO NOTHING;

-- 公司管理
INSERT INTO permissions (name, description) VALUES ('company:read', 'Allow reading company information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('company:create', 'Allow creating new companies') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('company:update', 'Allow updating company information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('company:delete', 'Allow deleting companies') ON CONFLICT (name) DO NOTHING;

-- 客戶管理
INSERT INTO permissions (name, description) VALUES ('customer:read', 'Allow reading customer information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('customer:create', 'Allow creating new customers') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('customer:update', 'Allow updating customer information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('customer:delete', 'Allow deleting customers') ON CONFLICT (name) DO NOTHING;

-- 選單管理
INSERT INTO permissions (name, description) VALUES ('menu:read', 'Allow reading menu information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('menu:create', 'Allow creating new menus') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('menu:update', 'Allow updating menu information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('menu:delete', 'Allow deleting menus') ON CONFLICT (name) DO NOTHING;

-- 角色選單管理
INSERT INTO permissions (name, description) VALUES ('role_menu:read', 'Allow reading role-menu relations') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('role_menu:create', 'Allow creating role-menu relations') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('role_menu:update', 'Allow updating role-menu relations') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('role_menu:delete', 'Allow deleting role-menu relations') ON CONFLICT (name) DO NOTHING;

-- 產品類別管理
INSERT INTO permissions (name, description) VALUES ('product_category:read', 'Allow reading product category information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_category:create', 'Allow creating new product categories') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_category:update', 'Allow updating product category information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_category:delete', 'Allow deleting product categories') ON CONFLICT (name) DO NOTHING;

-- 產品定義管理
INSERT INTO permissions (name, description) VALUES ('product_definition:read', 'Allow reading product definition information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_definition:create', 'Allow creating new product definitions') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_definition:update', 'Allow updating product definition information') ON CONFLICT (name) DO NOTHING;
INSERT INTO permissions (name, description) VALUES ('product_definition:delete', 'Allow deleting product definitions') ON CONFLICT (name) DO NOTHING;


-- 將所有權限賦予 'admin' 角色 (初始設定)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- 初始選單數據 (範例)
INSERT INTO menus (name, path, display_order) VALUES
('儀表板', '/dashboard', 10) ON CONFLICT (path) DO NOTHING,
('公司管理', '/dashboard/companies', 20) ON CONFLICT (path) DO NOTHING,
('客戶管理', '/dashboard/customers', 30) ON CONFLICT (path) DO NOTHING,
('產品定義', '/dashboard/product-definitions', 40) ON CONFLICT (path) DO NOTHING,
('帳戶管理', '/dashboard/accounts', 50) ON CONFLICT (path) DO NOTHING,
('選單管理', '/dashboard/menus', 60) ON CONFLICT (path) DO NOTHING,
('角色選單', '/dashboard/role-menus', 70) ON CONFLICT (path) DO NOTHING;

-- 將所有選單賦予 'admin' 角色 (初始設定)
INSERT INTO role_menus (role_id, menu_id)
SELECT r.id, m.id
FROM roles r, menus m
WHERE r.name = 'admin'
ON CONFLICT (role_id, menu_id) DO NOTHING;

-- 插入一個預設的管理員帳戶 (密碼請在運行 resetadmin 或手動雜湊後插入)
-- 這裡僅為範例，實際部署時應使用 resetadmin 工具或應用程式註冊流程
-- INSERT INTO accounts (username, password, role_id)
-- VALUES ('admin', 'hashed_password_from_bcrypt', (SELECT id FROM roles WHERE name = 'admin'));

-- 為了方便本地開發測試，可以在此處直接插入一個預設管理員帳戶和密碼。
-- 注意：生產環境絕不應這樣做！請使用 `resetadmin` 工具來設置管理員密碼。
-- 這裡的密碼 'password123' 已經是 bcrypt 雜湊後的 'hashed_password_for_password123'
-- 你可以使用 Go 程式碼生成雜湊：
-- go run -c 'import "golang.org/x/crypto/bcrypt"; import "fmt"; func main() { h, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost); fmt.Println(string(h)) }'
-- 替換以下行的 'hashed_password_for_password123' 為你生成的雜湊值
-- INSERT INTO accounts (username, password, role_id)
-- SELECT 'admin', '$2a$10$o.jK8M9Q4b5R6S7T8U9V0.jK8M9Q4b5R6S7T8U9V0e1M2N3O4P5Q6R7S8T9U0V1W2X3Y4Z5A6B7C8D9E0F1', id FROM roles WHERE name = 'admin'
-- ON CONFLICT (username) DO UPDATE SET password = EXCLUDED.password, updated_at = NOW();
