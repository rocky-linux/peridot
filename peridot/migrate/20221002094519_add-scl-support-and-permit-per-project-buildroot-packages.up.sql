ALTER TABLE
  IF EXISTS extra_package_options
ADD
  COLUMN IF NOT EXISTS depends_on text [] not null default array [] :: text [];

ALTER TABLE
  IF EXISTS projects
ADD
  COLUMN IF NOT EXISTS srpm_stage_packages text [] not null default array [] :: text [],
ADD
  COLUMN IF NOT EXISTS build_stage_packages text [] not null default array [] :: text [];
