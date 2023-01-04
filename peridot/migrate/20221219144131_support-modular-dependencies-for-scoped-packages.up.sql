ALTER TABLE
  IF EXISTS extra_package_options
ADD
  COLUMN IF NOT EXISTS enable_module text [] not null default array [] :: text [],
  COLUMN IF NOT EXISTS disable_module text [] not null default array [] :: text [];
