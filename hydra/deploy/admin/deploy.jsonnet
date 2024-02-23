local resfdeploy = import 'ci/resfdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local common = import 'hydra/deploy/common.jsonnet';

resfdeploy.new({
  name: 'hydra-admin',
  replicas: 1,
  dbname: 'hydra',
  backend: true,
  // Custom migration
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: common.sh_args($.dsn, '/usr/bin/hydra migrate sql -e --yes'),
  legacyDb: common.legacyDb,
  command: '/bin/sh',
  // We can use dangerous-force-http because we're using mTLS internally
  // and terminate TLS at ingress point.
  args: common.sh_args($.dsn, '/usr/bin/hydra serve admin'),
  image: common.image,
  tag: common.tag,
  dsn: {
    name: 'DSN',
    value: std.strReplace(db.dsn_legacy('hydra', false, 'hydra-admin'), 'postgresql://', 'postgres://') + "&max_conn_lifetime=5m",
  },
  requests: if kubernetes.prod() then {
    cpu: '0.2',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '2',
    memory: '8G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 4445,
      protocol: 'TCP',
    },
  ],
  health: {
    path: '/health/alive',
    port: 4445,
  },
  env: common.env + [$.dsn],
})
