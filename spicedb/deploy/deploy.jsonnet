local resfdeploy = import 'ci/resfdeploy.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local common = import 'spicedb/deploy/common.jsonnet';

resfdeploy.new({
  name: 'spicedb',
  replicas: 1,
  dbname: 'spicedb',
  backend: true,
  migrate: true,
  // Only create database
  migrate_command: ['/bin/sh'],
  migrate_args: common.sh_args('/usr/local/bin/spicedb migrate head --datastore-engine=postgres --datastore-conn-uri=$REAL_DSN'),
  legacyDb: common.legacyDb,
  command: '/bin/sh',
  // We can use dangerous-force-http because we're using mTLS internally
  // and terminate TLS at ingress point.
  args: common.sh_args('/usr/local/bin/spicedb serve --datastore-engine=postgres --datastore-conn-uri=$REAL_DSN'),
  image: common.image,
  tag: common.tag,
  dsn: common.dsn,
  internal_route_options: {
    headers: {
      request: {
        add: {
          'Authorization': 'Bearer %s' % common.env[0].value,
        }
      }
    }
  },
  requests: if kubernetes.prod() then {
    cpu: '0.2',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '1',
    memory: '2G',
  },
  ports: [
    {
      name: 'grpc',
      containerPort: 50051,
      protocol: 'TCP',
    },
    {
      name: 'internal',
      containerPort: 50053,
      protocol: 'TCP',
    },
    {
      name: 'prometheus',
      containerPort: 9090,
      protocol: 'TCP',
    },
  ],
  health: {
    grpc: true,
    port: 50051,
  },
  env: common.env,
})
