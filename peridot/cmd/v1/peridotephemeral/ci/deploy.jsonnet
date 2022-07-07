local bycdeploy = import 'ci/bycdeploy.jsonnet';
local db = import 'ci/db.jsonnet';
local kubernetes = import 'ci/kubernetes.jsonnet';
local temporal = import 'ci/temporal.jsonnet';
local utils = import 'ci/utils.jsonnet';

local site = std.extVar('site');

local provisionWorkerRole(metadata) = kubernetes.define_role_v2(metadata, 'provision-worker', [
  {
    apiGroups: [''],
    resources: ['pods', 'pods/log'],
    verbs: ['create', 'watch', 'get', 'delete'],
  }
]);

bycdeploy.new({
  name: 'peridotephemeral',
  replicas: if kubernetes.prod() then if site == 'extarches' then 5 else 10 else 1,
  dbname: 'peridot',
  backend: true,
  migrate: true,
  migrate_command: ['/bin/sh'],
  migrate_args: ['-c', 'exit 0'],
  legacyDb: true,
  command: '/bin/sh',
  args: ['-c', '/bundle/peridotephemeral*'],
  image: kubernetes.tag('peridotephemeral'),
  tag: kubernetes.version,
  dsn: {
    name: 'PERIDOTEPHEMERAL_DATABASE_URL',
    value: db.dsn_legacy('peridot', false, 'peridotephemeral'),
  },
  requests: if kubernetes.prod() then {
    cpu: '0.3',
    memory: '512M',
  },
  limits: if kubernetes.prod() then {
    cpu: '6',
    memory: '16G',
  },
  service_account_options: {
    annotations: {
      'eks.amazonaws.com/role-arn': 'arn:aws:iam::893168113496:role/peridot_k8s_role',
    }
  },
  ports: [
    {
      name: 'http',
      containerPort: 15201,
      protocol: 'TCP',
    },
  ],
  health: {
    port: 15201,
  },
  env: [
    {
      name: 'PERIDOTEPHEMERAL_PRODUCTION',
      value: if kubernetes.dev() then 'false' else 'true',
    },
    {
      name: 'PERIDOTEPHEMERAL_BUILDER_OCI_IMAGE_X86_64',
      value: kubernetes.tagVersion('peridotbuilder_amd64', $.tag),
    },
    {
      name: 'PERIDOTEPHEMERAL_BUILDER_OCI_IMAGE_AARCH64',
      value: kubernetes.tagVersion('peridotbuilder_arm64', $.tag),
    },
    {
      name: 'PERIDOTEPHEMERAL_BUILDER_OCI_IMAGE_S390X',
      value: kubernetes.tagVersion('peridotbuilder_s390x', $.tag),
    },
    {
      name: 'PERIDOTEPHEMERAL_BUILDER_OCI_IMAGE_PPC64LE',
      value: kubernetes.tagVersion('peridotbuilder_ppc64le', $.tag),
    },
    if kubernetes.prod() then {
      name: 'IMAGE_PULL_SECRET',
      value: 'registry',
    },
    if utils.local_image then {
      name: 'PERIDOTEPHEMERAL_S3_ENDPOINT',
      value: 'minio.default.svc.cluster.local:9000'
    },
    if utils.local_image then {
      name: 'PERIDOTEPHEMERAL_S3_DISABLE_SSL',
      value: 'true'
    },
    if utils.local_image then {
      name: 'PERIDOTEPHEMERAL_S3_FORCE_PATH_STYLE',
      value: 'true'
    },
    if utils.local_image then {
      name: 'PERIDOTEPHEMERAL_K8S_SUPPORTS_CROSS_PLATFORM_NO_AFFINITY',
      value: 'true'
    },
    if kubernetes.prod() then {
      name: 'PERIDOTEPHEMERAL_S3_REGION',
      value: 'us-east-2',
    },
    if kubernetes.prod() then {
      name: 'PERIDOTEPHEMERAL_S3_BUCKET',
      value: 'resf-peridot-prod',
    },
    if site == 'extarches' then {
      name: 'YUMREPOFS_HTTP_ENDPOINT_OVERRIDE',
      value: 'https://yumrepofs.build.resf.org',
    },
    if site == 'extarches' then {
      name: 'PERIDOTEPHEMERAL_TEMPORAL_HOSTPORT',
      value: 'temporal.corp.build.resf.org:443',
    } else temporal.kube_env('PERIDOTEPHEMERAL'),
    if site == 'extarches' then {
      name: 'PERIDOTEPHEMERAL_PROVISION_ONLY',
      value: 'true',
    },
    if site == 'extarches' then {
      name: 'PERIDOT_SITE',
      value: 'extarches',
    },
    if site == 'extarches' then {
      name: 'AWS_ACCESS_KEY_ID',
      valueFrom: true,
      secret: {
        name: 'aws',
        key: 'access-key-id',
      },
    },
    if site == 'extarches' then {
      name: 'AWS_SECRET_ACCESS_KEY',
      valueFrom: true,
      secret: {
        name: 'aws',
        key: 'secret-access-key',
      },
    },
    $.dsn,
  ],
  custom_job_items(metadata, extra): [
    provisionWorkerRole(metadata),
    kubernetes.bind_to_role_sa(provisionWorkerRole(metadata), extra.service_account_name)
  ],
})
