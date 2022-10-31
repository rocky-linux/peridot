local stage = std.extVar('stage');
local tag = std.extVar('tag');
local ociRegistry = std.extVar('oci_registry');
local ociRegistryRepo = std.extVar('oci_registry_repo');
local ociRegistryDocker = std.extVar('oci_registry_docker');
local ociNoNestedSupport = std.extVar('oci_registry_no_nested_support_in_2022_shame_on_you_aws') == 'true';
local site = std.extVar('site');
local arch = std.extVar('arch');
local localEnvironment = std.extVar('local_environment') == '1';

local utils = import 'ci/utils.jsonnet';

local helm_mode = utils.helm_mode;
local stage = utils.stage;
local user = utils.user;
local stageNoDash = utils.stage_no_dash;
local imagePullPolicy = if stageNoDash == 'dev' then 'Always' else 'IfNotPresent';

local defaultEnvs = [
  {
    name: 'RESF_ENV',
    value: stageNoDash,
  },
  {
    name: 'RESF_NS',
    valueFrom: true,
    field: 'metadata.namespace',
  },
  {
    name: 'RESF_FORCE_NS',
    value: if helm_mode then '{{ .Values.catalogForceNs | default !"!" }}' else '',
  },
  {
    name: 'RESF_SERVICE_ACCOUNT',
    valueFrom: true,
    field: 'spec.serviceAccountName',
  },
  {
    name: 'AWS_REGION',
    value: if helm_mode then '{{ .Values.awsRegion | default !"us-east-2!" }}' else 'us-east-2',
  },
  {
    name: 'LOCALSTACK_ENDPOINT',
    value: if utils.local_image then 'http://localstack.default.svc.cluster.local:4566' else '',
  }
];

local define_env(envsOrig) = std.filter(function(x) x != null, [
  if field != null then std.prune({
    name: field.name,
    value: if std.objectHas(field, 'value') then field.value,
    valueFrom: if std.objectHas(field, 'valueFrom') && field.valueFrom == true then {
      secretKeyRef: if std.objectHas(field, 'secret') then {
        name: field.secret.name,
        key: field.secret.key,
        optional: if std.objectHas(field.secret, 'optional') then field.secret.optional else false,
      },
      fieldRef: if std.objectHas(field, 'field') then {
        fieldPath: field.field,
      },
    },
  })
  for field in (envsOrig + defaultEnvs)
]);

local define_volumes(volumes) = [
  {
    name: vm.name,
    persistentVolumeClaim: if std.objectHas(vm, 'pvc') then {
      claimName: vm.name,
    },
    emptyDir: if std.objectHas(vm, 'emptyDir') then {},
    secret: if std.objectHas(vm, 'secret') then vm.secret,
    configMap: if std.objectHas(vm, 'configMap') then vm.configMap,
  }
  for vm in volumes
];

local define_volume_mounts(volumes) = [
  {
    name: vm.name,
    mountPath: vm.path,
  }
  for vm in volumes
];

local define_init_containers(initc_) = std.filter(function(x) x != null, [
  if initc != null && std.objectHas(initc, 'name') then {
    name: initc.name,
    image: initc.image,
    imagePullPolicy: imagePullPolicy,
    command: if std.objectHas(initc, 'command') then initc.command,
    args: if std.objectHas(initc, 'args') then initc.args,
    env: define_env(if std.objectHas(initc, 'env') then initc.env else []),
    volumeMounts: if std.objectHas(initc, 'volumes') && initc.volumes != null then define_volume_mounts(initc.volumes),
  }
  for initc in initc_
]);

local default_labels = {
  env: stageNoDash
};

local fix_metadata(metadata) = metadata {
  namespace: metadata.namespace,
};

local prod() = stage != '-dev';
local dev() = stage == '-dev';

{
  // For reference
  metadata: {
    name: 'empty',
    namespace: 'namespace',
    annotations: {},
  },

  // Namespace
  define_namespace(name, metadata={})::
    {
      apiVersion: 'v1',
      kind: 'Namespace',
      metadata: metadata {
        name: name,
      },
    },

  // Deployment
  define_deployment(metadataOrig, deporig)::
    local _ = std.assertEqual(true, std.objectHas(deporig, 'image'));
    local _ = std.assertEqual(true, std.objectHas(deporig, 'tag'));
    local metadata = fix_metadata(metadataOrig);

    local deployment = deporig {
      annotations: if !std.objectHas(deporig, 'annotations') then {} else deporig.annotations,
      labels: if !std.objectHas(deporig, 'labels') then default_labels else deporig.labels + default_labels,
      volumes: if !std.objectHas(deporig, 'volumes') then [] else deporig.volumes,
      imagePulLSecrets: if !std.objectHas(deporig, 'imagePullSecrets') then deporig.imagePullSecrets else deporig.imagePullSecrets,
      env: if !std.objectHas(deporig, 'env') then [] else deporig.env,
      ports: if !std.objectHas(deporig, 'ports') then [{ containerPort: 80, protocol: 'TCP' }] else deporig.ports,
      initContainers: if !std.objectHas(deporig, 'initContainers') then [] else deporig.initContainers,
      limits: if !std.objectHas(deporig, 'limits') || deporig.limits == null then { cpu: '0.1', memory: '256M' } else deporig.limits,
      requests: if !std.objectHas(deporig, 'requests') || deporig.requests == null then { cpu: '0.001', memory: '128M' } else deporig.requests,
    };

    {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: metadata {
        name: metadata.name + '-deployment',
      },
      spec: {
        revisionHistoryLimit: 15,
        selector: {
          matchLabels: {
            app: metadata.name,
            env: stageNoDash
          },
        },
        replicas: deployment.replicas,
        strategy: {
          type: 'RollingUpdate',
          rollingUpdate: {
            maxSurge: '300%',
            maxUnavailable: '0%',
          },
        },
        template: {
          metadata: {
            annotations: deployment.annotations,
            labels: deployment.labels {
              app: metadata.name,
              env: stageNoDash,
              version: deployment.tag,
            },
          },
          spec: {
            automountServiceAccountToken: true,
            serviceAccountName: if std.objectHas(deployment, 'serviceAccount') then deployment.serviceAccount,
            initContainers: if std.objectHas(deployment, 'initContainers') && deployment.initContainers != null then define_init_containers(deployment.initContainers),
            securityContext: {
              fsGroup: 1000,
            },
            containers: [
              {
                image: deployment.image + (if ociNoNestedSupport then '-' else ':') + deployment.tag,
                imagePullPolicy: imagePullPolicy,
                name: metadata.name,
                command: if std.objectHas(deployment, 'command') then deployment.command else null,
                args: if std.objectHas(deployment, 'args') then deployment.args else null,
                ports: deployment.ports,
                env: if std.objectHas(deployment, 'env') then define_env(deployment.env) else [],
                volumeMounts: if std.objectHas(deployment, 'volumes') && deployment.volumes != null then define_volume_mounts(deployment.volumes),
                securityContext: {
                  runAsGroup: if std.objectHas(deployment, 'fsGroup') then deployment.fsGroup else null,
                  runAsUser: if std.objectHas(deployment, 'fsUser') then deployment.fsUser else null,
                },
                resources: {
                  limits: deployment.limits,
                  requests: deployment.requests,
                },
                readinessProbe: if std.objectHas(deployment, 'health') && deployment.health != null then {
                  httpGet: if !std.objectHas(deployment.health, 'grpc') || !deployment.health.grpc then {
                    path: if std.objectHas(deployment.health, 'path') then deployment.health.path else '/_/healthz',
                    port: deployment.health.port,
                    httpHeaders: [
                      {
                        name: 'resf-internal-req',
                        value: 'yes',
                      },
                    ],
                  },
                  exec: if std.objectHas(deployment.health, 'grpc') && deployment.health.grpc then {
                    command: ["grpc_health_probe", "-connect-timeout=4s", "-v", "-addr=localhost:"+deployment.health.port],
                  },
                  initialDelaySeconds: if std.objectHas(deployment.health, 'initialDelaySeconds') then deployment.health.initialDelaySeconds else 1,
                  periodSeconds: if std.objectHas(deployment.health, 'periodSeconds') then deployment.health.periodSeconds else 3,
                  timeoutSeconds: if std.objectHas(deployment.health, 'timeoutSeconds') then deployment.health.timeoutSeconds else 5,
                  successThreshold: if std.objectHas(deployment.health, 'successThreshold') then deployment.health.successThreshold else 1,
                  failureThreshold: if std.objectHas(deployment.health, 'failureThreshold') then deployment.health.failureThreshold else 30,
                } else if std.objectHas(deployment, 'health_tcp') && deployment.health_tcp != null then {
                  tcpSocket: {
                    port: deployment.health_tcp.port,
                  },
                  initialDelaySeconds: if std.objectHas(deployment.health, 'initialDelaySeconds') then deployment.health.initialDelaySeconds else 5,
                  periodSeconds: if std.objectHas(deployment.health, 'periodSeconds') then deployment.health.periodSeconds else 5,
                },
              },
            ],
            affinity: if !std.objectHas(deployment, 'no_anti_affinity') || !deployment.no_anti_affinity then {
              podAntiAffinity: {
                preferredDuringSchedulingIgnoredDuringExecution: [
                  {
                    weight: 99,
                    podAffinityTerm: {
                      labelSelector: {
                        matchExpressions: [
                          {
                            key: 'app',
                            operator: 'In',
                            values: [
                              metadata.name,
                            ],
                          },
                        ],
                      },
                      topologyKey: 'kubernetes.io/hostname',
                    },
                  },
                  {
                    weight: 100,
                    podAffinityTerm: {
                      labelSelector: {
                        matchExpressions: [
                          {
                            key: 'app',
                            operator: 'In',
                            values: [
                              metadata.name,
                            ],
                          },
                        ],
                      },
                      topologyKey: 'failure-domain.beta.kubernetes.io/zone',
                    },
                  },
                ],
              },
            },
            restartPolicy: 'Always',
            imagePullSecrets: if std.objectHas(deployment, 'imagePullSecrets') && deployment.imagePullSecrets != null then if std.type(deployment.imagePullSecrets) == 'string' then deployment.imagePullSecrets else [
              {
                name: secret,
              }
              for secret in deployment.imagePullSecrets
            ],
            volumes: if std.objectHas(deployment, 'volumes') && deployment.volumes != null then define_volumes(deployment.volumes),
          },
        },
      },
    },

  // Ingress
  define_ingress(metadataOrig, host, srvName=null, path='/', port=80)::
    local metadata = fix_metadata(metadataOrig);

    {
      apiVersion: 'networking.k8s.io/v1',
      kind: 'Ingress',
      metadata: metadata {
        name: metadata.name + '-ingress',
      },
      spec: {
        rules: [{
          host: host,
          http: {
            paths: [
              {
                path: path,
                pathType: 'Prefix',
                backend: {
                  service: {
                    name: if srvName != null then srvName else metadata.name + '-service',
                    port: {
                      number: port,
                    }
                  }
                },
              },
            ],
          },
        }],
      } + (if !helm_mode then {} else {
        tls: [{
          hosts: [
            host,
          ],
          secretName: metadata.name + '-tls',
        }],
      }),
    },

  // Service
  define_service(metadataOrig, externalPort=80, internalPort=80, protocol='TCP', portName='http', selector='', env='canary')::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: metadata {
        name: metadata.name + '-service',
      },
      spec: {
        type: 'ClusterIP',
        ports: [{
          name: portName,
          port: externalPort,
          protocol: protocol,
          targetPort: internalPort,
        }] + (if portName == 'http' && externalPort != 80 then [{
          name: portName + "-80",
          port: 80,
          protocol: protocol,
          targetPort: internalPort,
        }] else []),
        selector: {
          app: if selector != '' then selector else metadata.name,
          env: env,
        },
      },
    },

  // Virtual Service
  define_virtual_service(metadataOrig, spec)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'VirtualService',
      metadata: metadata {
        name: metadata.name + '-vs',
      },
      spec: spec,
    },

  // Destination rule
  define_destination_rule(metadataOrig, spec)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'DestinationRule',
      metadata: metadata {
        name: metadata.name + '-dsr',
      },
      spec: spec,
    },

  // Service entry
  define_service_entry(metadataOrig, hosts, ports, resolution, location)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'networking.istio.io/v1alpha3',
      kind: 'ServiceEntry',
      metadata: metadata {
        name: metadata.name + '-se',
      },
      spec: {
        hosts: hosts,
        ports: ports,
        resolution: resolution,
        location: location,
      },
    },

  // Job
  define_job(metadataOrig, joborig)::
    local metadata = fix_metadata(metadataOrig);
    local job = joborig {
      env: if !std.objectHas(joborig, 'env') then [] else joborig.env,
      labels: if !std.objectHas(joborig, 'labels') then {} else joborig.labels,
      annotations: if !std.objectHas(joborig, 'annotations') then {} else joborig.annotations,
      initContainers: if !std.objectHas(joborig, 'initContainers') then [] else joborig.initContainers,
      volumes: if !std.objectHas(joborig, 'volumes') then [] else joborig.volumes,
      args: if !std.objectHas(joborig, 'args') then [] else joborig.args,
    };

    local name = metadata.name + '-job';

    {
      apiVersion: 'batch/v1',
      kind: 'Job',
      metadata: metadata {
        name: name,
      },
      spec: {
        ttlSecondsAfterFinished: 120,
        template: {
          metadata: {
            labels: job.labels,
            annotations: job.annotations,
          },
          spec: {
            automountServiceAccountToken: true,
            serviceAccountName: if std.objectHas(job, 'serviceAccount') then job.serviceAccount,
            imagePullSecrets: if std.objectHas(job, 'imagePullSecrets') && job.imagePullSecrets != null then if std.type(job.imagePullSecrets) == 'string' then job.imagePullSecrets else [
              {
                name: secret,
              }
              for secret in job.imagePullSecrets
            ],
            initContainers: define_init_containers(job.initContainers),
            containers: [{
              name: name,
              image: job.image + (if ociNoNestedSupport then '-' else ':') + job.tag,
              command: if std.objectHas(job, 'command') then job.command else null,
              args: job.args,
              env: define_env(job.env),
              volumeMounts: if std.objectHas(job, 'volumes') && job.volumes != null then define_volume_mounts(job.volumes),
            }],
            restartPolicy: if std.objectHas(job, 'restartPolicy') then job.restartPolicy else 'Never',
            volumes: if std.objectHas(job, 'volumes') && job.volumes != null then define_volumes(job.volumes),
          },
        },
      },
    },

  // ServiceAccount
  define_service_account(metadataOrig)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: metadata {
        name: metadata.name,
      },
    },

  // Role
  define_role(metadataOrig, rules)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'Role',
      metadata: metadata {
        name: metadata.name + '-role',
      },
      rules: rules,
    },

  define_role_v2(metadataOrig, name, rules)::
    $.define_role(metadataOrig { name: '%s-%s' % [metadataOrig.name, name] }, rules),

  // ClusterRole
  define_cluster_role(metadataOrig, rules)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: metadata {
        name: metadata.name + '-clusterrole',
      },
      rules: rules,
    },

  define_cluster_role_v2(metadataOrig, name, rules)::
    $.define_cluster_role(metadataOrig { name: '%s-%s' % [metadataOrig.name, name] }, rules),

  // RoleBinding
  define_role_binding(metadataOrig, roleName, subjects)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'RoleBinding',
      metadata: metadata {
        name: metadata.name + '-rolebinding',
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: roleName,
      },
      subjects: subjects,
    },
  bind_to_role_sa(role, serviceAccount)::
    $.define_role_binding(role.metadata, role.metadata.name, [{
      kind: 'ServiceAccount',
      name: serviceAccount,
      namespace: role.metadata.namespace,
    }]),

  // ClusterRoleBinding
  define_cluster_role_binding(metadataOrig, roleName, subjects)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: metadata {
        name: metadata.name + '-clusterrolebinding',
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: roleName,
      },
      subjects: subjects,
    },
  bind_to_cluster_role_sa(role, serviceAccount)::
    $.define_cluster_role_binding(role.metadata, role.metadata.name, [{
      kind: 'ServiceAccount',
      name: serviceAccount,
      namespace: role.metadata.namespace,
    }]),

  // PersistentVolumeClaim
  define_persistent_volume_claim(metadataOrig, storage, access_mode='ReadWriteOnce')::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'v1',
      kind: 'PersistentVolumeClaim',
      metadata: metadata {
        name: metadata.name + '-pvc',
      },
      spec: {
        accessModes: [access_mode],
        resources: {
          requests: {
            storage: storage,
          },
        },
      },
    },

  // ConfigMap
  define_config_map(metadataOrig, data)::
    local metadata = fix_metadata(metadataOrig);
    {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: metadata {
        name: metadata.name + '-cm',
      },
      data: data,
    },

  chown_vm(name, path, id, volumes)::
    {
      name: 'chown-vm-' + name,
      image: 'alpine:3.9.3',
      command: [
        'chown',
        '-R',
        '%d:%d' % [id, id],
        path,
      ],
      volumes: volumes,
    },

  istio_labels()::
    {
      'istio-injection': 'enabled',
    },

  tag(name, extra=false)::
    '%s/%s%s%s%s' % [
      std.strReplace(ociRegistry, 'host.docker.internal.local', 'registry'),
      ociRegistryRepo,
      if ociNoNestedSupport then ':' else '/',
      name,
      if !extra then (if (arch != 'amd64' && !localEnvironment) then '_'+arch else '') else '',
    ],

  tagVersion(name, version)::
    '%s%s%s' % [$.tag(name, true), (if ociNoNestedSupport then '-' else ':'), version],

  fix_metadata: fix_metadata,

  prod: prod,

  dev: dev,

  version: tag,
}
