local resfdeploy = import 'ci/resfdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';
local s3 = import 'ci/s3.jsonnet';

resfdeploy.new({
  name: 'peridotadmin',
  helm_strip_prefix: 'PERIDOTADMIN_',
  replicas: 1,
  dbname: 'peridot',
  backend: true,
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bundle/peridotadmin',
  image: kubernetes.tag('peridotadmin'),
  tag: kubernetes.version,
  dsn: {
    name: 'PERIDOTADMIN_DATABASE_URL',
    value: db.dsn_legacy('peridot', false, 'peridotadmin'),
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
      'eks.amazonaws.com/role-arn': if utils.helm_mode then '{{ .Values.awsRoleArn | default !"!" }}' else 'arn:aws:iam::893168113496:role/peridot_k8s_role',
    }
  },
  ports: [
    {
      name: 'http',
      containerPort: 15012,
      protocol: 'TCP',
    },
    {
      name: 'grpc',
      containerPort: 15013,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 15012,
  },
  env: [
    {
      name: 'PERIDOTADMIN_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    {
      name: 'HYDRA_PUBLIC_HTTP_ENDPOINT_OVERRIDE',
      value: if utils.helm_mode then '{{ .Values.hydraPublicEndpoint | default !"!" }}' else '',
    },
    {
      name: 'HYDRA_ADMIN_HTTP_ENDPOINT_OVERRIDE',
      value: if utils.helm_mode then '{{ .Values.hydraAdminEndpoint | default !"!" }}' else '',
    },
    {
      name: 'SPICEDB_GRPC_ENDPOINT_OVERRIDE',
      value: if utils.helm_mode then '{{ .Values.spicedbEndpoint | default !"!" }}' else '',
    },
    $.dsn,
  ] + s3.kube_env('PERIDOTADMIN') + temporal.kube_env('PERIDOTADMIN'),
})
