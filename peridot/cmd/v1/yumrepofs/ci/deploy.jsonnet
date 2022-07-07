local bycdeploy = import 'ci/bycdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local utils = import 'ci/utils.jsonnet';

bycdeploy.new({
  name: 'yumrepofs',
  replicas: if kubernetes.prod() then 3 else 1,
  dbname: 'peridot',
  backend: true,
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bundle/yumrepofs',
  image: kubernetes.tag('yumrepofs'),
  tag: kubernetes.version,
  service_account_options: {
    annotations: {
      'eks.amazonaws.com/role-arn': 'arn:aws:iam::893168113496:role/peridot_k8s_role',
    }
  },
  dsn: {
    name: 'YUMREPOFS_DATABASE_URL',
    value: db.dsn_legacy('peridot', false, 'yumrepofs'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.2',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '2',
    // We may need extra memory while serving big packages, rare but may happen.
    // Setting memory higher is a safety measure
    memory: '24G',
  },
  ports: [
    {
      name: 'http',
      containerPort: 45002,
      protocol: 'TCP',
      expose: true,
    },
    {
      name: 'grpc',
      containerPort: 45003,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 45002,
  },
  env: [
    {
      name: 'YUMREPOFS_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    if utils.local_image then {
      name: 'YUMREPOFS_S3_ENDPOINT',
      value: 'minio.default.svc.cluster.local:9000'
    },
    if utils.local_image then {
      name: 'YUMREPOFS_S3_DISABLE_SSL',
      value: 'true'
    },
    if utils.local_image then {
      name: 'YUMREPOFS_S3_FORCE_PATH_STYLE',
      value: 'true'
    },
    if kubernetes.prod() then {
      name: 'YUMREPOFS_S3_REGION',
      value: 'us-east-2',
    },
    if kubernetes.prod() then {
      name: 'YUMREPOFS_S3_BUCKET',
      value: 'resf-peridot-prod',
    },
    {
      name: 'YUMREPOFS_S3_ASSUME_ROLE',
      value: 'arn:aws:iam::893168113496:role/peridot_s3_upload_role',
    },
    $.dsn,
  ],
})
