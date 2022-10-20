ALTER TABLE
  IF EXISTS extra_package_options DROP COLUMN IF EXISTS depends_on;

ALTER TABLE
  IF EXISTS projects DROP COLUMN IF EXISTS srpm_stage_packages,
  DROP COLUMN IF EXISTS build_stage_packages;
