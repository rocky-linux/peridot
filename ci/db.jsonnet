local stage = std.extVar('stage');
local origUser = std.extVar('user');
local domainUser = std.extVar('domain_user');
local ociRegistry = std.extVar('oci_registry');
local utils = import 'ci/utils.jsonnet';

local user = if domainUser != 'user-orig' then domainUser else origUser;

{
  host()::
    'prod-db-cockroachdb-public.cockroachdb.svc.cluster.local',

  port()::
    '26257',

  host_port()::
    '%s:%s' % [$.host(), $.port()],

  cert(name)::
    '/cockroach-certs/client.%s.crt' % name,

  key(name)::
    '/cockroach-certs/client.%s.key' % name,

  ca()::
    '/cockroach-certs/ca.crt',

  label()::
    { 'cockroachdb-client': 'true' },

  obj_label()::
    { labels: $.label() },

  staged_name(name)::
    '%s%s' % [name, std.strReplace(stage, '-', '')],

  dsn_raw(name, password)::
    local staged_name = $.staged_name(name);
    'postgresql://%s%s@cockroachdb-public.cockroachdb.svc.cluster.local:26257' %
    [staged_name, if password then ':byc' else ':REPLACEME'] +
    '/%s?sslmode=require&sslcert=/cockroach-certs/client.%s.crt' %
    [staged_name, staged_name] +
    '&sslkey=/cockroach-certs/client.%s.key&sslrootcert=/cockroach-certs/ca.crt' %
    [staged_name],

  dsn_legacy(name, no_add=false, ns=null)::
      local staged_name = $.staged_name(name);
      if utils.local_image then 'postgresql://postgres:postgres@postgres-postgresql.default.svc.cluster.local:5432/%s?sslmode=disable' % staged_name
      else 'postgresql://%s%s:REPLACEME@resf-peridot-dev.ctxqgglmfofx.us-east-2.rds.amazonaws.com:5432' %
      [if !no_add then (if stage == '-dev' then '%s-dev' % user else if ns != null then ns else name)+'-' else '', if no_add then name else staged_name] +
      '/%s?sslmode=disable' %
      [if no_add then name else staged_name],

  dsn(name)::
    $.dsn_raw(name, false),
}

