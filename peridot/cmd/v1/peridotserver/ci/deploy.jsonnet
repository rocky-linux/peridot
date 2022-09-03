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
  service_account_options: {
    annotations: {
      'eks.amazonaws.com/role-arn': 'arn:aws:iam::893168113496:role/peridot_k8s_role',
    }
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
    if utils.local_image then {
      name: 'PERIDOT_S3_ENDPOINT',
      value: 'minio.default.svc.cluster.local:9000'
    },
    if utils.local_image then {
      name: 'PERIDOT_S3_DISABLE_SSL',
      value: 'true'
    },
    if utils.local_image then {
      name: 'PERIDOT_S3_FORCE_PATH_STYLE',
      value: 'true'
    },
    $.dsn,
  ] + temporal.kube_env('PERIDOT'),
})
