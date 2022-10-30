local RESFDEPLOY = import 'ci/RESFDEPLOY.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

RESFDEPLOY.new({
  name: 'apollo',
  replicas: 1,
  dbname: 'apollo',
  backend: true,
  migrate: true,
  legacyDb: true,
  command: '/bundle/apollo',
  image: kubernetes.tag('apollo'),
  tag: kubernetes.version,
  dsn: {
    name: 'APOLLO_DATABASE_URL',
    value: db.dsn_legacy('apollo'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.5',
    memory: '512M',
  },
  ports: [
    {
      name: 'http',
      containerPort: 9100,
      protocol: 'TCP',
      expose: true,
    },
    {
      name: 'grpc',
      containerPort: 9101,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 9100,
  },
  env: [
    {
      name: 'APOLLO_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    $.dsn,
  ] + temporal.kube_env('APOLLO'),
})
