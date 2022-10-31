local stage = std.extVar('stage');
local origUser = std.extVar('user');
local domainUser = std.extVar('domain_user');
local ociRegistry = std.extVar('oci_registry');
local utils = import 'ci/utils.jsonnet';

local user = if domainUser != 'user-orig' then domainUser else origUser;
local default_host_port = 'resf-peridot-dev.ctxqgglmfofx.us-east-2.rds.amazonaws.com:5432';

local host_port = if !utils.helm_mode then default_host_port else '{{ .Values.postgresqlHostPort }}';

{
  staged_name(name)::
    '%s%s' % [name, if utils.helm_mode then '{{ template !"resf.stage!" . }}!!' else std.strReplace(stage, '-', '')],

  dsn_inner(name, no_add=false, ns=null)::
      local staged_name = $.staged_name(name);
      if utils.local_image then 'postgresql://postgres:postgres@postgres-postgresql.default.svc.cluster.local:5432/%s?sslmode=disable' % staged_name
      else 'postgresql://%s%s:REPLACEME@%s' %
      [if !no_add then (if utils.helm_mode then '!!{{ .Release.Namespace }}-' else (if stage == '-dev' then '%s-dev' % user else if ns != null then ns else name)+'-') else '', if no_add then name else staged_name, host_port] +
      '/%s?sslmode=disable' %
      [if no_add then name else staged_name],

  dsn(name, no_add=false, ns=null)::
      local res = $.dsn_inner(name, no_add, ns);
      if utils.helm_mode then '!!{{ if .Values.databaseUrl }}{{ .Values.databaseUrl }}{{ else }}%s{{end}}!!' % res else res,

  dsn_legacy(name, no_add=false, ns=null)::
    $.dsn(name, no_add, ns),
}

