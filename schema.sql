-- Database schema for Pijar App with Midtrans integration

-- Create users table
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

-- Create products table
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    product_id INTEGER NOT NULL REFERENCES products(id),
    amount INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    order_id VARCHAR(100) NOT NULL UNIQUE,
    payment_url TEXT,
    midtrans_id VARCHAR(100),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create orders table (if needed for more complex order management)
CREATE TABLE IF NOT EXISTS orders (
    id VARCHAR(100) PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id),
    gross_amount INTEGER NOT NULL,
    customer VARCHAR(100) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create order_items table (for tracking items in an order)
CREATE TABLE IF NOT EXISTS order_items (
    id SERIAL PRIMARY KEY,
    order_id VARCHAR(100) NOT NULL REFERENCES orders(id),
    item_id VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    price INTEGER NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    monthly_subscription INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_transactions_user_id ON transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_product_id ON transactions(product_id);
CREATE INDEX IF NOT EXISTS idx_transactions_order_id ON transactions(order_id);
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON order_items(order_id);

-- Insert sample product data
INSERT INTO products (name, description, price) VALUES
('Basic Plan', 'Basic subscription plan', 50000),
('Premium Plan', 'Premium subscription with more features', 100000),
('Enterprise Plan', 'Full-featured enterprise subscription', 200000)
ON CONFLICT (id) DO NOTHING;

-- Insert sample admin user
INSERT INTO users (name, email, password_hash, birth_year, phone, role) VALUES
('Admin User', 'admin@example.com', '$2a$10$VQYFCiCye2edww8xhjnbpOp9uhM0QSJHSw/bCNekp/VoeEABxTrXe', 1990, '08123456789', 'ADMIN')
ON CONFLICT (email) DO NOTHING;


========
-- DDL Statements

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create topics table
CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    preference VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create articles table
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    source VARCHAR(255) NOT NULL,
    topic_id INTEGER NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create journals table
CREATE TABLE IF NOT EXISTS journals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    judul VARCHAR(255) NOT NULL,
    isi TEXT NOT NULL,
    perasaan VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create coach_sessions table
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

-- Create conversation_contexts table
CREATE TABLE IF NOT EXISTS conversation_contexts (
    session_id VARCHAR(36) PRIMARY KEY,
    context JSONB NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES coach_sessions(session_id) ON DELETE CASCADE
);

-- Create user_goals table
CREATE TABLE IF NOT EXISTS user_goals (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    task TEXT NOT NULL,
    articles_to_read INTEGER[] NOT NULL,
    completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create goal_progress table
CREATE TABLE IF NOT EXISTS goal_progress (
    id SERIAL PRIMARY KEY,
    goal_id INTEGER NOT NULL REFERENCES user_goals(id) ON DELETE CASCADE,
    article_id INTEGER NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    date_assigned TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(goal_id, article_id)
);

-- Create indexes for better performance
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

-- Goal progress indexes
CREATE INDEX IF NOT EXISTS idx_goal_progress_goal_id ON goal_progress(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_article_id ON goal_progress(article_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_completed ON goal_progress(completed);

-- Create indexes for better performance
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

-- Goal progress indexes
CREATE INDEX IF NOT EXISTS idx_goal_progress_goal_id ON goal_progress(goal_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_article_id ON goal_progress(article_id);
CREATE INDEX IF NOT EXISTS idx_goal_progress_completed ON goal_progress(completed);

======
-- Pijar Application Database Schema

-- Drop tables if they exist (for clean setup)
DROP TABLE IF EXISTS user_goals_progress;
DROP TABLE IF EXISTS user_goals;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS articles;

-- Create users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create articles table (for reading materials)
CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    author VARCHAR(100),
    published_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    category VARCHAR(50),
    estimated_read_time INT -- in minutes
);

-- Create user_goals table
CREATE TABLE user_goals (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL,
    task TEXT NOT NULL,
    articles_to_read INT[] DEFAULT '{}',
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create user_goals_progress table to track progress on articles
CREATE TABLE user_goals_progress (
    id SERIAL PRIMARY KEY,
    id_goals INT NOT NULL,
    id_article INT NOT NULL,
    date_completed TIMESTAMP,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (id_goals) REFERENCES user_goals(id) ON DELETE CASCADE,
    UNIQUE(id_goals, id_article) -- Ensure each article is only tracked once per goal
);

-- Create indexes for better query performance
CREATE INDEX idx_user_goals_user_id ON user_goals(user_id);
CREATE INDEX idx_user_goals_progress_goal_id ON user_goals_progress(id_goals);
CREATE INDEX idx_user_goals_progress_completed ON user_goals_progress(completed);

-- Sample data insertion for testing

-- Insert sample users
INSERT INTO users (username, email, password) VALUES 
('john_doe', 'john@example.com', 'hashed_password_here'),
('jane_smith', 'jane@example.com', 'hashed_password_here');

-- Insert sample articles
INSERT INTO articles (title, content, author, category, estimated_read_time) VALUES
('Getting Started with Go', 'This article covers the basics of Go programming language...', 'Go Team', 'Programming', 15),
('Effective Time Management', 'Learn how to manage your time effectively...', 'Productivity Expert', 'Productivity', 10),
('Introduction to PostgreSQL', 'A comprehensive guide to PostgreSQL database...', 'DB Admin', 'Database', 20);

-- Insert sample goals
INSERT INTO user_goals (user_id, title, task, articles_to_read) VALUES
(1, 'Learn Go Programming', 'Complete basic Go tutorials and build a simple API', '{1, 3}'),
(2, 'Improve Productivity', 'Implement new time management techniques', '{2}');
