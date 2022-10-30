create table reboot_suggested_packages
(
  created_at timestamp default now() not null,

  name       text unique             not null
)
