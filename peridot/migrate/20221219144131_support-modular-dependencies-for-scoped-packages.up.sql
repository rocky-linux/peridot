ALTER TABLE IF EXISTS extra_package_options ADD COLUMN IF NOT EXISTS enable_module text [] NOT NULL DEFAULT ARRAY []::text [],
ADD COLUMN IF NOT EXISTS disable_module text [] NOT NULL DEFAULT ARRAY []::text [];
