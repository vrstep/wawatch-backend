CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    role VARCHAR(255),
    profile_picture VARCHAR(255) DEFAULT 'default.jpg'
);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE TABLE IF NOT EXISTS anime_caches (
    id INT PRIMARY KEY, -- Anilist ID, not auto-incrementing
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    title VARCHAR(255),
    cover_image TEXT,
    format VARCHAR(50),
    total_episodes INT
);
CREATE INDEX IF NOT EXISTS idx_anime_caches_deleted_at ON anime_caches(deleted_at);
CREATE INDEX IF NOT EXISTS idx_anime_caches_title ON anime_caches(title);

CREATE TABLE IF NOT EXISTS user_anime_lists (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    user_id INT NOT NULL,
    anime_external_id INT NOT NULL,
    status VARCHAR(20),
    score INT,
    progress INT DEFAULT 0,
    start_date TIMESTAMPTZ,
    end_date TIMESTAMPTZ,
    notes TEXT,
    rewatch_count INT DEFAULT 0,
    CONSTRAINT fk_user_anime_lists_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_user_anime_lists_anime FOREIGN KEY (anime_external_id) REFERENCES anime_caches(id) ON DELETE CASCADE -- Assuming you want this constraint
);
CREATE INDEX IF NOT EXISTS idx_user_anime_lists_deleted_at ON user_anime_lists(deleted_at);
CREATE INDEX IF NOT EXISTS idx_user_anime_lists_user_id ON user_anime_lists(user_id);
CREATE INDEX IF NOT EXISTS idx_user_anime_lists_anime_external_id ON user_anime_lists(anime_external_id);
CREATE INDEX IF NOT EXISTS idx_user_anime_lists_status ON user_anime_lists(status);

-- Ensure the uuid-ossp extension is enabled if you haven't already in PostgreSQL
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS watch_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    anime_id INT NOT NULL,
    provider_name VARCHAR(255) NOT NULL,
    provider_url TEXT,
    region VARCHAR(2),
    is_sub BOOLEAN DEFAULT false,
    is_dub BOOLEAN DEFAULT false,
    last_updated TIMESTAMPTZ,
    CONSTRAINT fk_watch_providers_anime FOREIGN KEY (anime_id) REFERENCES anime_caches(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_watch_providers_deleted_at ON watch_providers(deleted_at);