local kubernetes = import 'ci/kubernetes.jsonnet';
local db = import 'ci/db.jsonnet';
local utils = import 'ci/utils.jsonnet';

local tag = std.extVar('tag');

local DSN = db.dsn('hydra');
local authn = if kubernetes.prod() then 'https://id.build.resf.org' else 'https://id-dev.internal.pdev.resf.localhost';

{
  image: 'oryd/hydra',
  tag: 'v2.0.3',
  legacyDb: true,
  env: [
    {
      name: 'URLS_SELF_ISSUER',
      value: if kubernetes.prod() then 'https://hdr.build.resf.org/' else 'https://hdr-dev.internal.pdev.resf.localhost',
    },
    {
      name: 'URLS_SELF_PUBLIC',
      value: if kubernetes.prod() then 'https://hdr.build.resf.org/' else 'https://hdr-dev.internal.pdev.resf.localhost',
    },
    {
      name: 'URLS_LOGIN',
      value: '%s/login' % authn
    },
    {
      name: 'URLS_CONSENT',
      value: '%s/consent' % authn
    },
    {
      name: 'URLS_LOGOUT',
      value: '%s/logout' % authn
    },
    {
      name: 'URLS_ERROR',
      value: '%s/error' % authn
    },
    {
      name: 'URLS_POST_LOGOUT_REDIRECT',
      value: 'https://rockylinux.org'
    },
    {
      name: 'SERVE_TLS_ALLOW_TERMINATION_FROM',
      value: '127.0.0.1/32,172.39.0.0/16,100.96.0.0/24'
    },
    {
      name: 'LOG_LEAK_SENSITIVE_VALUES',
      value: if utils.local_image then 'true' else 'false'
    },
    {
      name: 'SECRETS_SYSTEM',
      valueFrom: true,
      secret: {
        name: 'hydra',
        key: 'system-secret',
      }
    },
    {
      name: 'SECRETS_COOKIE',
      valueFrom: true,
      secret: {
        name: 'hydra',
        key: 'cookie-secret',
      }
    },
  ],
  sh_args(dsn, cmd): [
    '-c',
    'export REAL_DSN=`echo $%s | sed -e "s/REPLACEME/${DATABASE_PASSWORD}/g"%s`; DSN=$REAL_DSN %s' % [dsn.name, if $.legacyDb then '' else ' | sed -e "s/postgresql/cockroachdb/g"', cmd],
  ]
}
