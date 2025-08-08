-- Create subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
                                             id SERIAL PRIMARY KEY,
                                             service_name VARCHAR(255) NOT NULL,
    price INTEGER NOT NULL CHECK (price > 0),
    user_id UUID NOT NULL,
    start_date VARCHAR(7) NOT NULL, -- Format: MM-YYYY
    end_date VARCHAR(7), -- Optional, Format: MM-YYYY
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
                             );

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_service_name ON subscriptions(service_name);
CREATE INDEX IF NOT EXISTS idx_subscriptions_start_date ON subscriptions(start_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_end_date ON subscriptions(end_date);
CREATE INDEX IF NOT EXISTS idx_subscriptions_deleted_at ON subscriptions(deleted_at);

-- Add constraints
ALTER TABLE subscriptions
    ADD CONSTRAINT chk_start_date_format
        CHECK (start_date ~ '^(0[1-9]|1[0-2])-[0-9]{4}$');

ALTER TABLE subscriptions
    ADD CONSTRAINT chk_end_date_format
        CHECK (end_date IS NULL OR end_date ~ '^(0[1-9]|1[0-2])-[0-9]{4}$');