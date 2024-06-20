local kubernetes = import 'ci/kubernetes.jsonnet';
local db = import 'ci/db.jsonnet';
local tag = std.extVar('tag');

local DSN = db.dsn('hydra');

{
  image: 'quay.io/peridot/spicedb',
  tag: 'v0.3.29',
  legacyDb: true,
  dsn: {
    name: 'DSN',
    value: std.strReplace(db.dsn_legacy('spicedb'), 'postgresql://', 'postgres://'),
  },
  env: [
    {
      name: 'SPICEDB_GRPC_PRESHARED_KEY',
      // This may be insecure, but it's a necessary evil.
      // todo(mustafa): Evaluate whether we can use a gRPC proxy instead
      value: 'iKeNRY7ZMZaksFO0mX8uMFCzL8Ayzcq1',
      /*valueFrom: true,
      secret: {
        name: 'spicedb',
        key: 'grpc-preshared-key',
      }*/
    },
    $.dsn
  ],
  sh_args(cmd): [
    '-c',
    'export REAL_DSN=`echo $%s | sed -e "s/REPLACEME/${DATABASE_PASSWORD}/g"%s`; DSN=$REAL_DSN %s' % [$.dsn.name, if $.legacyDb then '' else ' | sed -e "s/postgresql/cockroachdb/g"', cmd],
  ]
}
