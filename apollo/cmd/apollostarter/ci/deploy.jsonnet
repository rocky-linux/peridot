local RESFDEPLOY = import 'ci/RESFDEPLOY.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

local site = std.extVar('site');

RESFDEPLOY.new({
  name: 'apollostarter',
  replicas: 1,
  dbname: 'apollo',
  backend: true,
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bundle/apollostarter',
  image: kubernetes.tag('apollostarter'),
  tag: kubernetes.version,
  dsn: {
    name: 'APOLLOSTARTER_DATABASE_URL',
    value: db.dsn_legacy('apollo', false, 'apollostarter'),
  },
  requests: if kubernetes.prod() then {
    cpu: '1',
    memory: '2G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 31209,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 31209,
  },
  env: [
    {
      name: 'APOLLOSTARTER_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    $.dsn,
  ] + temporal.kube_env('APOLLOSTARTER'),
})
