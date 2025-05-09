-- ========================================================================
-- DDL (DATA DEFINITION LANGUAGE) STATEMENTS
-- ========================================================================

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    birth_year INTEGER NOT NULL,
    phone VARCHAR(20) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'USER',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


-- Products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    product_id INTEGER NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    amount INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_id VARCHAR(100) NOT NULL UNIQUE,
    payment_url TEXT,
    midtrans_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Orders table (for more complex order management)
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(100) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    gross_amount INTEGER NOT NULL,
    customer VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Order_items table (for tracking items in an order)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(100) NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    item_id VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    price INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    monthly_subscription INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Topics table
CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Articles table
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    source VARCHAR(255) NOT NULL,
    author VARCHAR(100),
    category VARCHAR(50),
    estimated_read_time INTEGER,
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Journals table
CREATE TABLE IF NOT EXISTS journals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    judul VARCHAR(255) NOT NULL,
    isi TEXT NOT NULL,
    perasaan VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Coach_sessions table
CREATE TABLE IF NOT EXISTS coach_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(36) NOT NULL UNIQUE,
    timestamp TIMESTAMP NOT NULL,
    user_input TEXT,
    ai_response TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Conversation_contexts table
CREATE TABLE IF NOT EXISTS conversation_contexts (
    session_id VARCHAR(36) PRIMARY KEY,
    context JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES coach_sessions(session_id) ON DELETE CASCADE
);

-- User_goals table
CREATE TABLE IF NOT EXISTS user_goals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    task TEXT NOT NULL,
    articles_to_read INTEGER[] DEFAULT '{}',
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User_goals_progress table to track progress on articles
CREATE TABLE IF NOT EXISTS user_goals_progress (
    id SERIAL PRIMARY KEY,
    id_goals INTEGER NOT NULL,
    id_article INTEGER NOT NULL,
    date_completed TIMESTAMP,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (id_goals) REFERENCES user_goals(id) ON DELETE CASCADE,
    FOREIGN KEY (id_article) REFERENCES articles(id) ON DELETE CASCADE,
    UNIQUE(id_goals, id_article) -- Ensure each article is only tracked once per goal
);

-- ========================================================================
-- INDEXES
-- ========================================================================

-- Transactions indexes
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_product_id ON transactions(product_id);
CREATE INDEX IF NOT EXISTS idx_transactions_order_id ON transactions(order_id);

-- Orders indexes
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);

-- Order_items indexes
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Users indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Topics indexes
CREATE INDEX IF NOT EXISTS idx_topics_user_id ON topics(user_id);
CREATE INDEX IF NOT EXISTS idx_topics_preference ON topics(preference);

-- Articles indexes
CREATE INDEX IF NOT EXISTS idx_articles_topic_id ON articles(topic_id);
CREATE INDEX IF NOT EXISTS idx_articles_title ON articles(title);

-- Journals indexes
CREATE INDEX IF NOT EXISTS idx_journals_user_id ON journals(user_id);
CREATE INDEX IF NOT EXISTS idx_journals_created_at ON journals(created_at);

-- Coach sessions indexes
CREATE INDEX IF NOT EXISTS idx_coach_sessions_user_id ON coach_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_coach_sessions_session_id ON coach_sessions(session_id);
CREATE INDEX IF NOT EXISTS idx_coach_sessions_timestamp ON coach_sessions(timestamp);

-- Conversation contexts indexes
CREATE INDEX IF NOT EXISTS idx_conversation_contexts_session_id ON conversation_contexts(session_id);

-- User goals indexes
CREATE INDEX IF NOT EXISTS idx_user_goals_user_id ON user_goals(user_id);
CREATE INDEX IF NOT EXISTS idx_user_goals_created_at ON user_goals(created_at);

-- User goals progress indexes
CREATE INDEX IF NOT EXISTS idx_user_goals_progress_goal_id ON user_goals_progress(id_goals);
CREATE INDEX IF NOT EXISTS idx_user_goals_progress_article_id ON user_goals_progress(id_article);
CREATE INDEX IF NOT EXISTS idx_user_goals_progress_completed ON user_goals_progress(completed);

-- ========================================================================
-- DML (DATA MANIPULATION LANGUAGE) STATEMENTS
-- ========================================================================

-- Sample products data
INSERT INTO products (name, description, price) VALUES
('Basic Plan', 'Basic subscription plan', 50000),
('Premium Plan', 'Premium subscription with more features', 100000),
('Enterprise Plan', 'Full-featured enterprise subscription', 200000)
ON CONFLICT (id) DO NOTHING;

-- Sample admin user
INSERT INTO users (name, email, password_hash, birth_year, phone, role) 
VALUES (
    'Admin User', 
    'admin1@example.com', 
    '$2a$10$VQYFCiCye2edww8xhjnbpOp9uhM0QSJHSw/bCNekp/VoeEABxTrXe', 
    1990, 
    '08123456789', 
    'ADMIN'),
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (name, email, password_hash, birth_year, phone, role) 
VALUES (
    'New Admin 22' ,
    'admin22@example.com',
    crypt('securepassword', gen_salt('bf', 10)),
    1995,
    '08123456780',
    'ADMIN');

-- Sample topics (assumes user_id 1 exists)
INSERT INTO topics (user_id, preference) VALUES 
(1, 'Programming'),
(1, 'Productivity'),
(1, 'Database')
ON CONFLICT DO NOTHING;

-- Sample articles (assumes topic_ids 1, 2, 3 exist)
INSERT INTO articles (title, content, source, author, category, estimated_read_time, topic_id) VALUES
('Getting Started with Go', 'This article covers the basics of Go programming language. Go is an open source programming language designed at Google that makes it easy to build simple, reliable, and efficient software. It combines the development speed of working in a dynamic language like Python with the performance and safety of a compiled language like C or C++. In this comprehensive guide, we will explore the fundamentals of Go, including its syntax, data types, control structures, functions, methods, interfaces, and concurrency patterns. By the end of this article, you will have a solid understanding of Go programming and be able to write your own Go applications.', 'Go Blog', 'Go Team', 'Programming', 15, 1),
('Effective Time Management', 'Learn how to manage your time effectively with proven techniques and strategies. Time management is the process of planning and exercising conscious control of time spent on specific activities, especially to increase effectiveness, efficiency, and productivity. Poor time management can result in missed deadlines, inadequate work quality, higher stress levels, and poor professional reputation. This article explores various time management techniques including the Pomodoro Technique, Eisenhower Box, time blocking, and the 80/20 rule. We will also discuss how to identify and eliminate time-wasting activities, set appropriate boundaries, and develop sustainable productivity habits that can transform your work and personal life.', 'Productivity Blog', 'Productivity Expert', 'Productivity', 10, 2),
('Introduction to PostgreSQL', 'A comprehensive guide to PostgreSQL database system. PostgreSQL is a powerful, open-source object-relational database system with over 30 years of active development that has earned it a strong reputation for reliability, feature robustness, and performance. This guide covers PostgreSQL installation, basic SQL commands, advanced features like stored procedures, triggers, views, and extensions, as well as performance optimization techniques. You will learn how to design efficient database schemas, implement proper indexing strategies, and scale your PostgreSQL deployment. The article also explores PostgreSQL''s unique features compared to other database systems, including its extensive data type support, sophisticated locking mechanism, and robust transaction processing capabilities. Whether you are a beginner or an experienced database professional, this guide will help you leverage the full power of PostgreSQL for your applications.', 'Database Journal', 'DB Admin', 'Database', 20, 3)
ON CONFLICT DO NOTHING;

-- Sample user goals (assumes user_id 1 and 2 exist)
INSERT INTO user_goals (user_id, title, task, articles_to_read) VALUES
(1, 'Learn Go Programming', 'Complete basic Go tutorials and build a simple API', '{1, 3}'),
(2, 'Improve Productivity', 'Implement new time management techniques', '{2}')
ON CONFLICT DO NOTHING;