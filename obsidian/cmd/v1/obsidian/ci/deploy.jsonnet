local RESFDEPLOY = import 'ci/RESFDEPLOY.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';

RESFDEPLOY.new({
  name: 'obsidian',
  dbname: 'obsidian',
  backend: true,
  migrate: true,
  legacyDb: true,
  command: '/bundle/obsidian',
  image: kubernetes.tag('obsidian'),
  tag: kubernetes.version,
  dsn: {
    name: 'OBSIDIAN_DATABASE_URL',
    value: db.dsn_legacy('obsidian'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.2',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '0.3',
    memory: '1G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 26002,
      protocol: 'TCP',
      expose: true,
    },
    {
      name: 'grpc',
      containerPort: 26003,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 26002,
  },
  env: [
    {
      name: 'OBSIDIAN_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    $.dsn,
  ],
})
