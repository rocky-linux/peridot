local bycdeploy = import 'ci/bycdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

bycdeploy.new({
  name: 'keykeeper',
  replicas: if kubernetes.prod() then 20 else 3,
  dbname: 'peridot',
  backend: true,
  migrate: true,
  fsUser: 0,
  fsGroup: 0,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bundle/keykeeper',
  image: kubernetes.tag('keykeeper'),
  tag: kubernetes.version,
  service_account_options: {
    annotations: {
      'eks.amazonaws.com/role-arn': 'arn:aws:iam::893168113496:role/peridot_k8s_role',
    }
  },
  dsn: {
    name: 'KEYKEEPER_DATABASE_URL',
    value: db.dsn_legacy('peridot', false, 'keykeeper'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.25',
    memory: '1G',
  },
  limits: if kubernetes.prod() then {
    cpu: '20',
    memory: '128G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 46002,
      protocol: 'TCP',
      expose: kubernetes.dev(),
    },
    {
      name: 'grpc',
      containerPort: 46003,
      protocol: 'TCP',
      expose: true,
    },
  ],
  health: {
    port: 46002,
  },
  env: [
    {
      name: 'KEYKEEPER_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    if utils.local_image then {
      name: 'KEYKEEPER_S3_ENDPOINT',
      value: 'minio.default.svc.cluster.local:9000'
    },
    if utils.local_image then {
      name: 'KEYKEEPER_S3_DISABLE_SSL',
      value: 'true'
    },
    if utils.local_image then {
      name: 'KEYKEEPER_S3_FORCE_PATH_STYLE',
      value: 'true'
    },
    {
      name: 'KEYKEEPER_AWSSM_PREFIX',
      value: 'keykeeper_',
    },
    if kubernetes.prod() then {
      name: 'KEYKEEPER_S3_REGION',
      value: 'us-east-2',
    },
    if kubernetes.prod() then {
      name: 'KEYKEEPER_S3_BUCKET',
      value: 'resf-peridot-prod',
    },
    $.dsn,
  ] + temporal.kube_env('KEYKEEPER'),
})
