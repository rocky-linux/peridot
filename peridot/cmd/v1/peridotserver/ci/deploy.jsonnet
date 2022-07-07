local bycdeploy = import 'ci/bycdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

bycdeploy.new({
  name: 'peridotserver',
  replicas: if kubernetes.prod() then 5 else 1,
  dbname: 'peridot',
  backend: true,
  migrate: true,
  legacyDb: true,
  command: '/bundle/peridotserver',
  image: kubernetes.tag('peridotserver'),
  tag: kubernetes.version,
  dsn: {
    name: 'PERIDOT_DATABASE_URL',
    value: db.dsn_legacy('peridot', false, 'peridotserver'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.2',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '2',
    memory: '12G',
  } else {
    cpu: '2',
    memory: '10G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 15002,
      protocol: 'TCP',
      expose: true,
    },
    {
      name: 'grpc',
      containerPort: 15003,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 15002,
  },
  env: [
    {
      name: 'PERIDOT_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    temporal.kube_env('PERIDOT'),
    $.dsn,
  ],
})
