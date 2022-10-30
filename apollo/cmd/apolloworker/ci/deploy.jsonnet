local resfdeploy = import 'ci/resfdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

local site = std.extVar('site');

resfdeploy.new({
  name: 'apolloworker',
  replicas: 1,
  dbname: 'apollo',
  backend: true,
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bundle/apolloworker',
  image: kubernetes.tag('apolloworker'),
  tag: kubernetes.version,
  dsn: {
    name: 'APOLLOWORKER_DATABASE_URL',
    value: db.dsn_legacy('apollo', false, 'apolloworker'),
  },
  requests: if kubernetes.prod() then {
    cpu: '1',
    memory: '2G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 29209,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 29209,
  },
  env: [
    {
      name: 'APOLLOWORKER_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    $.dsn,
  ] + temporal.kube_env('APOLLOWORKER'),
})
