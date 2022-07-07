local stage = std.extVar('stage');
local origUser = std.extVar('user');
local domainUser = std.extVar('domain_user');
local ociRegistry = std.extVar('oci_registry');
local ociRegistryRepo = std.extVar('oci_registry_repo');
local registry_secret = std.extVar('registry_secret');

local user = if domainUser != 'user-orig' then domainUser else origUser;

local stageNoDash = std.strReplace(stage, '-', '');

local kubernetes = import 'ci/kubernetes.jsonnet';
local db = import 'ci/db.jsonnet';
local mappings = import 'ci/mappings.jsonnet';
local utils = import 'ci/utils.jsonnet';

local labels = {
  labels: db.label() + kubernetes.istio_labels(),
};

{
  new(info)::
    local metadata = {
      name: info.name,
      namespace: if stageNoDash == 'dev' then '%s-dev' % user else if std.objectHas(info, 'namespace') then info.namespace else info.name,
    };
    local fixed = kubernetes.fix_metadata(metadata);
    local vshost(srv) = '%s-service.%s.svc.cluster.local' % [srv.name, fixed.namespace];
    local infolabels = if info.backend then labels else { labels: kubernetes.istio_labels() };
    local dbname = (if std.objectHas(info, 'dbname') then info.dbname else info.name);
    local env = if std.objectHas(info, 'env') then info.env else [];
    local sa_name = '%s-%s-serviceaccount' % [stageNoDash, fixed.name];

    local extra_info = {
      service_account_name: sa_name,
    };

    local envs = [stageNoDash];

    local services = if std.objectHas(info, 'services') then info.services else
      [{ name: '%s-%s-%s' % [metadata.name, port.name, env], port: port.containerPort, portName: port.name, expose: if std.objectHas(port, 'expose') then port.expose else false } for env in envs for port in info.ports];

    local nssa = '001-ns-sa.yaml';
    local migrate = '002-migrate.yaml';
    local deployment = '003-deployment.yaml';
    local svcVsDr = '004-svc-vs-dr.yaml';
    local custom = '005-custom.yaml';

    local legacyDb = if std.objectHas(info, 'legacyDb') then info.legacyDb else false;

    local dbPassEnv = {
      name: 'DATABASE_PASSWORD',
      valueFrom: !utils.local_image,
      value: if utils.local_image then 'postgres',
      secret: if !utils.local_image then {
        name: '%s-database-password' % db.staged_name(dbname),
        key: 'password',
      },
    };

    local shouldSecureEndpoint(srv) = if mappings.get_env_from_svc(srv.name) == 'prod' && mappings.is_external(srv.name) then false
                                      else if mappings.should_expose_all(srv.name) then false
                                      else if utils.local_image then false
                                      else if !std.objectHas(srv, 'expose') || !srv.expose then false
                                      else true;
    local imagePullSecrets = if registry_secret != 'none' then [registry_secret] else [];

    {
      [nssa]: std.manifestYamlStream([
        kubernetes.define_namespace(metadata.namespace, infolabels),
        kubernetes.define_service_account(metadata {
          name: '%s-%s' % [stageNoDash, fixed.name],
        } + if std.objectHas(info, 'service_account_options') then info.service_account_options else {},
        ),
      ]),
      [if std.objectHas(info, 'migrate') && info.migrate == true then migrate else null]:
        std.manifestYamlStream([
          kubernetes.define_service_account(metadata {
            name: 'init-db-%s-%s' % [fixed.name, stageNoDash],
          }),
          kubernetes.define_role_binding(metadata, metadata.name + '-role', [{
            kind: 'ServiceAccount',
            name: 'init-db-%s-%s-serviceaccount' % [fixed.name, stageNoDash],
            namespace: metadata.namespace,
          }]),
          kubernetes.define_cluster_role_binding(metadata, metadata.name + '-clusterrole', [{
            kind: 'ServiceAccount',
            name: 'init-db-%s-%s-serviceaccount' % [fixed.name, stageNoDash],
            namespace: metadata.namespace,
          }]),
          kubernetes.define_role(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, stageNoDash],
            },
            [{
              apiGroups: [''],
              resources: ['secrets'],
              verbs: ['create', 'get'],
            }]
          ),
          kubernetes.define_cluster_role(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, stageNoDash],
            },
            [{
              apiGroups: [''],
              resources: ['secrets'],
              verbs: ['create', 'get'],
            }]
          ),
          kubernetes.define_role_binding(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, stageNoDash],
            },
            'init-db-%s-%s-role' % [fixed.name, stageNoDash],
            [{
              kind: 'ServiceAccount',
              name: 'init-db-%s-%s-serviceaccount' % [fixed.name, stageNoDash],
              namespace: fixed.namespace,
            }],
          ),
          kubernetes.define_cluster_role_binding(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, stageNoDash],
            },
            'init-db-%s-%s-clusterrole' % [fixed.name, stageNoDash],
            [{
              kind: 'ServiceAccount',
              name: 'init-db-%s-%s-serviceaccount' % [fixed.name, stageNoDash],
              namespace: fixed.namespace,
            }],
          ),
          if !legacyDb then kubernetes.define_job(metadata { name: 'request-cert' }, kubernetes.request_cdb_certs('initdb%s' % stageNoDash) + {
            serviceAccount: '%s-%s-serviceaccount' % [stageNoDash, fixed.name],
          }),
          if info.migrate == true && dbname != '' then kubernetes.define_job(metadata { name: info.name + '-migrate' }, {
            image: if std.objectHas(info, 'migrate_image') && info.migrate_image != null then info.migrate_image else info.image,
            tag: if std.objectHas(info, 'migrate_tag') && info.migrate_tag != null then info.migrate_tag else info.tag,
            command: if std.objectHas(info, 'migrate_command') && info.migrate_command != null then info.migrate_command else ['/bin/sh'],
            serviceAccount: 'init-db-%s-%s-serviceaccount' % [fixed.name, stageNoDash],
            imagePullSecrets: imagePullSecrets,
            args: if std.objectHas(info, 'migrate_args') && info.migrate_args != null then info.migrate_args else [
              '-c',
              'export REAL_DSN=`echo $%s | sed -e "s/REPLACEME/${DATABASE_PASSWORD}/g"%s`; /usr/bin/migrate -source file:///migrations -database $REAL_DSN up' % [info.dsn.name, if legacyDb then '' else ' | sed -e "s/postgresql/cockroachdb/g"'],
            ],
            volumes: (if std.objectHas(info, 'volumes') then info.volumes(metadata) else []) + (if !legacyDb then kubernetes.request_cdb_certs_volumes() else []),
            initContainers: [
              if !legacyDb then kubernetes.request_cdb_certs('%s%s' % [metadata.name, stageNoDash]) + {
                serviceAccount: '%s-%s-serviceaccount' % [stageNoDash, fixed.name],
              },
              {
                name: 'initdb',
                image: 'quay.io/peridot/initdb:v0.1.4',
                command: ['/bin/sh'],
                args: ['-c', '/bundle/initdb*'],
                volumes: if !legacyDb then kubernetes.request_cdb_certs_volumes(),
                env: [
                  {
                    name: 'INITDB_TARGET_DB',
                    value: db.staged_name(dbname),
                  },
                  {
                    name: 'INITDB_PRODUCTION',
                    value: 'true',
                  },
                  {
                    name: 'INITDB_DATABASE_URL',
                    value: if legacyDb then db.dsn_legacy('postgres', true) else db.dsn('initdb'),
                  },
                ],
              },
            ],
            env: [
              dbPassEnv,
              info.dsn,
            ],
            annotations: {
              'sidecar.istio.io/inject': 'false',
            },
          }) else {},
        ]),
      [deployment]: std.manifestYamlStream([
        kubernetes.define_deployment(
          metadata,
          {
            replicas: if std.objectHas(info, 'replicas') then info.replicas else 1,
            image: info.image,
            tag: info.tag,
            command: if std.objectHas(info, 'command') then [info.command] else null,
            fsGroup: if std.objectHas(info, 'fsGroup') then info.fsGroup else null,
            fsUser: if std.objectHas(info, 'fsUser') then info.fsUser else null,
            imagePullSecrets: imagePullSecrets,
            labels: db.label(),
            annotations: if std.objectHas(info, 'annotations') then info.annotations,
            initContainers: if !legacyDb && info.backend then [kubernetes.request_cdb_certs('%s%s' % [metadata.name, stageNoDash]) + {
              serviceAccount: '%s-%s-serviceaccount' % [stageNoDash, fixed.name],
            }],
            volumes: (if std.objectHas(info, 'volumes') then info.volumes(metadata) else []) + (if !legacyDb then kubernetes.request_cdb_certs_volumes() else []),
            ports: std.map(function(x) x { expose: null, external: null }, info.ports),
            health: if std.objectHas(info, 'health') then info.health,
            env: env + (if dbname != '' && info.backend then ([dbPassEnv]) else []) + [
              {
                name: 'SELF_IDENTITY',
                value: 'spiffe://cluster.local/ns/%s/sa/%s-%s-serviceaccount' % [fixed.namespace, stageNoDash, fixed.name],
              },
            ] + [
              if std.objectHas(srv, 'expose') && srv.expose then {
                name: '%s_PUBLIC_URL' % [std.asciiUpper(std.strReplace(std.strReplace(srv.name, stage, ''), '-', '_'))],
                value: 'https://%s' % mappings.get(srv.name, user),
              } else null,
            for srv in services],
            limits: if std.objectHas(info, 'limits') then info.limits,
            requests: if std.objectHas(info, 'requests') then info.requests,
            args: if std.objectHas(info, 'args') then info.args else [],
            serviceAccount: '%s-%s-serviceaccount' % [stageNoDash, fixed.name],
          },
        ),
      ]),
      [svcVsDr]:
        std.manifestYamlStream(
          [kubernetes.define_service(metadata { name: srv.name }, srv.port, srv.port, portName=srv.portName, selector=metadata.name, env=mappings.get_env_from_svc(srv.name)) for srv in services] +
          [kubernetes.define_virtual_service(metadata { name: srv.name + '-internal' }, {
            hosts: [vshost(srv)],
            gateways: [],
            http: [
              {
                route: [{
                  destination: {
                    host: vshost(srv),
                    subset: mappings.get_env_from_svc(srv.name),
                    port: {
                      number: srv.port,
                    },
                  },
                } + (if std.objectHas(info, 'internal_route_options') then info.internal_route_options else {})],
              },
            ],
          },) for srv in services] +
          [if std.objectHas(srv, 'expose') && srv.expose then kubernetes.define_virtual_service(
            metadata {
              name: srv.name,
              annotations: {
                'external-dns.alpha.kubernetes.io/target': if mappings.is_external(srv.name) then 'ingress.build.resf.org' else 'ingress-internal.build.resf.org',
              },
            },
            {
              hosts: [mappings.get(srv.name, user)],
              gateways: if mappings.is_external(srv.name) then ['istio-system/base-gateway-public'] else ['istio-system/base-gateway-confidential'],
              http: [
                {
                  route: [{
                    destination: {
                      host: vshost(srv),
                      subset: mappings.get_env_from_svc(srv.name),
                      port: {
                        number: srv.port,
                      },
                    },
                  } + (if std.objectHas(info, 'external_route_options') then info.external_route_options else {})],
                },
              ],
            }
          ) else null for srv in services] +
          [{
              apiVersion: 'security.istio.io/v1beta1',
              kind: 'RequestAuthentication',
              metadata: metadata {
                name: srv.name,
              },
              spec: {
                selector: {
                  matchLabels: {
                    app: metadata.name,
                    env: mappings.get_env_from_svc(srv.name),
                  },
                },
                // todo(mustafa): Introduce ObsidianProxy to support BeyondCorp
                jwtRules: if shouldSecureEndpoint(srv) then [{
                  issuer: 'https://cloud.google.com/iap',
                  jwksUri: 'https://www.gstatic.com/iap/verify/public_key-jwk',
                  fromHeaders: [{ name: 'x-goog-iap-jwt-assertion' }],
                }] else [],
            },
          } for srv in services] +
          [{
            apiVersion: 'security.istio.io/v1beta1',
            kind: 'AuthorizationPolicy',
            metadata: metadata {
              name: srv.name,
            },
            spec: {
              selector: {
                matchLabels: {
                  app: metadata.name,
                  env: mappings.get_env_from_svc(srv.name),
                },
              },
              action: 'ALLOW',
              rules: [(if shouldSecureEndpoint(srv) then {
                from: [],
              } else {}) + {
                to: [{
                  operation: {
                    ports: [std.toString(srv.port)]
                  }
                }]
              }],
            },
          } for srv in services] +
          [kubernetes.define_destination_rule(metadata { name: srv.name }, {
            host: vshost(srv),
            trafficPolicy: {
              tls: {
                mode: 'ISTIO_MUTUAL',
              },
            },
            subsets: [
              {
                name: mappings.get_env_from_svc(srv.name),
                labels: {
                  app: metadata.name,
                  env: mappings.get_env_from_svc(srv.name),
                },
              },
            ],
          },) for srv in services]
        ),
      [if std.objectHas(info, 'custom_job_items') then custom else null]:
        std.manifestYamlStream(if std.objectHas(info, 'custom_job_items') then info.custom_job_items(metadata, extra_info) else [{}]),
    },
}
