local ociRegistry = std.extVar('oci_registry');
local ociRegistryRepo = std.extVar('oci_registry_repo');
local registry_secret = std.extVar('registry_secret');

local kubernetes = import 'ci/kubernetes.jsonnet';
local db = import 'ci/db.jsonnet';
local mappings = import 'ci/mappings.jsonnet';
local utils = import 'ci/utils.jsonnet';

local helm_mode = utils.helm_mode;
local stage = utils.stage;
local user = utils.user;
local stageNoDash = utils.stage_no_dash;

local slugify_ = function (x) std.asciiLower(std.substr(x, 0, 1)) + std.substr(x, 1, std.length(x)-1);
local slugify = function (name, extra_remove, str) slugify_(std.join('', [std.asciiUpper(std.substr(x, 0, 1)) + std.asciiLower(std.substr(x, 1, std.length(x)-1)) for x in std.split(std.strReplace(std.strReplace(str, std.asciiUpper(name)+'_', ''), extra_remove, ''), '_')]));

// Can be used to add common labels or annotations
local labels = {
  labels: kubernetes.istio_labels(),
};

// We're using a helper manifestYamlStream function to fix some general issues with it for Helm templates manually.
// Currently Helm functions should use !" instead of " only for strings.
// If a value doesn't start with a Helm bracket but ends with one, then end the value with !! (and the opposite for start).
local manifestYamlStream = function (value, indent_array_in_object=false, c_document_end=false, quote_keys=false)
  std.strReplace(std.strReplace(std.strReplace(std.strReplace(std.strReplace(std.manifestYamlStream(std.filter(function (x) x != null, value), indent_array_in_object, c_document_end, quote_keys), '!\\"', '"'), '"{{', '{{'), '}}"', '}}'), '}}!!', '}}'), '!!{{', '{{');

{
  user():: user,
  new(info)::
    local metadata_init = {
      name: info.name,
      namespace: if helm_mode then '{{ .Release.Namespace }}' else (if stageNoDash == 'dev' then '%s-dev' % user else if std.objectHas(info, 'namespace') then info.namespace else info.name),
    };
    local default_labels_all = {
      'app.kubernetes.io/name': if helm_mode then '{{ template !"%s.name!" . }}' % info.name else info.name,
    };
    local default_labels_helm = if helm_mode then {
      'helm.sh/chart': '{{ template !"%s.chart!" . }}' % info.name,
      'app.kubernetes.io/managed-by': '{{ .Release.Service }}',
      'app.kubernetes.io/instance': '{{ .Release.Name }}',
      'app.kubernetes.io/version': info.tag,
    } else {};
    local default_labels = default_labels_all + default_labels_helm;
    local metadata = metadata_init + { labels: default_labels };
    local fixed = kubernetes.fix_metadata(metadata);
    local vshost(srv) = '%s-service.%s.svc.cluster.local' % [srv.name, fixed.namespace];
    local infolabels = if info.backend then labels else { labels: kubernetes.istio_labels() };
    local dbname = (if std.objectHas(info, 'dbname') then info.dbname else info.name);
    local env = std.filter(function (x) x != null, [if x != null then if (!std.endsWith(x.name, 'DATABASE_URL') && std.objectHas(x, 'value') && x.value != null) && std.findSubstr('{{', x.value) == null then x {
      value: if helm_mode then '{{ .Values.%s | default !"%s!"%s }}' % [slugify(info.name, if std.objectHas(info, 'helm_strip_prefix') then info.helm_strip_prefix else ' ', x.name), x.value, if x.value == 'true' || x.value == 'false' then ' | quote' else ''] else x.value,
    } else x for x in (if std.objectHas(info, 'env') then info.env else [])]);
    local sa_default = fixed.name;
    local sa_name = if helm_mode then '{{ .Values.serviceAccountName | default !"%s!" }}' % [fixed.name] else sa_default;

    local envs = [stageNoDash];

    local disableMetrics = std.objectHas(info, 'disableMetrics') && info.disableMetrics;
    local ports = (if std.objectHas(info, 'ports') then info.ports else []) + (if disableMetrics then [] else [{
      name: 'metrics',
      containerPort: 7332,
      protocol: 'TCP',
    }]);
    local services = if std.objectHas(info, 'services') then info.services else
      [{ name: '%s-%s-%s' % [metadata.name, port.name, env], port: port.containerPort, portName: port.name, expose: if std.objectHas(port, 'expose') then port.expose else false } for env in envs for port in ports];

    local file_yaml_prefix = if helm_mode then 'helm-' else '';
    local nssa = '%s001-ns-sa.yaml' % file_yaml_prefix;
    local migrate = '%s002-migrate.yaml' % file_yaml_prefix;
    local deployment = '%s003-deployment.yaml' % file_yaml_prefix;
    local svcVsDr = '%s004-svc-vs-dr.yaml' % file_yaml_prefix;
    local custom = '%s005-custom.yaml' % file_yaml_prefix;

    local dbPassEnv = {
      name: 'DATABASE_PASSWORD',
      valueFrom: !utils.local_image,
      value: if utils.local_image then 'postgres',
      secret: if !utils.local_image then {
        name: '%s-database-password' % db.staged_name(dbname),
        key: 'password',
        optional: if utils.helm_mode then '{{ if .Values.databaseUrl }}true{{ else }}false{{ end }}' else false,
      },
    };

    local ingress_annotations = {
      'kubernetes.io/tls-acme': 'true',
      'cert-manager.io/cluster-issuer': if utils.helm_mode then '!!{{ if .Values.overrideClusterIssuer }}{{ .Values.overrideClusterIssuer }}{{ else }}letsencrypt-{{ template !"resf.stage!" . }}{{ end }}!!' else 'letsencrypt-staging',
    } + (if utils.local_image || !info.backend then {
      'konghq.com/https-redirect-status-code': '301',
    } else {});

    // Helm mode doesn't need this as the deployer/operator should configure it themselves
    local shouldSecureEndpoint(srv) = if helm_mode then false else (if mappings.get_env_from_svc(srv.name) == 'prod' && mappings.is_external(srv.name) then false
                                      else if mappings.should_expose_all(srv.name) then false
                                      else if utils.local_image then false
                                      else if !std.objectHas(srv, 'expose') || !srv.expose then false
                                      else true);
    local imagePullSecrets = if helm_mode then '{{ if .Values.imagePullSecrets }}[{{ range .Values.imagePullSecrets }}{ name: {{.}} },{{ end }}]{{ else }}null{{end}}' else (if registry_secret != 'none' then [registry_secret] else []);
    local migrate_image = if std.objectHas(info, 'migrate_image') && info.migrate_image != null then info.migrate_image else info.image;
    local migrate_tag = if std.objectHas(info, 'migrate_tag') && info.migrate_tag != null then info.migrate_tag else info.tag;
    local stage_in_resource = if helm_mode then '%s!!' % stage else stage;
    local image = if helm_mode then '{{ ((.Values.image).repository) | default !"%s!" }}' % info.image else info.image;
    local tag = if helm_mode then '{{ ((.Values.image).tag) | default !"%s!" }}' % info.tag else info.tag;

    local extra_info = {
      service_account_name: sa_name,
      imagePullSecrets: imagePullSecrets,
      image: image,
      tag: tag,
    };
    local istio_mode = true; #if helm_mode then false else if utils.local_image then false else true;

    {
      [nssa]: (if helm_mode then '{{ if not .Values.serviceAccountName }}\n' else '') + manifestYamlStream([
        // disable namespace creation in helm mode
        if !helm_mode then kubernetes.define_namespace(metadata.namespace, infolabels + { annotations: { 'linkerd.io/inject': 'enabled' } }),
        kubernetes.define_service_account(
          metadata {
            name: fixed.name,
          } + if std.objectHas(info, 'service_account_options') then info.service_account_options else {}
        ),
      ]) + (if helm_mode then '{{ end }}' else ''),
      [if std.objectHas(info, 'migrate') && info.migrate == true then migrate else null]:
        manifestYamlStream([
          kubernetes.define_service_account(metadata {
            name: 'init-db-%s' % [fixed.name],
          }),
          kubernetes.define_role(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, fixed.namespace],
              namespace: 'initdb%s' % stage_in_resource,
            },
            [{
              apiGroups: [''],
              resources: ['secrets'],
              verbs: ['get'],
            }]
          ),
          kubernetes.define_role(
            metadata {
              name: 'init-db-%s' % [fixed.name],
            },
            [{
              apiGroups: [''],
              resources: ['secrets'],
              verbs: ['create', 'get'],
            }]
          ),
          kubernetes.define_role_binding(
            metadata {
              name: 'init-db-%s-%s' % [fixed.name, fixed.namespace],
              namespace: 'initdb%s' % stage_in_resource,
            },
            'init-db-%s-%s-role' % [fixed.name, fixed.namespace],
            [{
              kind: 'ServiceAccount',
              name: 'init-db-%s' % [fixed.name],
              namespace: fixed.namespace,
            }],
          ),
          kubernetes.define_role_binding(
            metadata {
              name: 'init-db-%s' % [fixed.name],
            },
            'init-db-%s-role' % [fixed.name],
            [{
              kind: 'ServiceAccount',
              name: 'init-db-%s' % [fixed.name],
              namespace: fixed.namespace,
            }],
          ),
          if info.migrate == true && dbname != '' then kubernetes.define_job(
            metadata {
              name: info.name + '-migrate',
              annotations: (if helm_mode then {
                'helm.sh/hook': 'post-install,post-upgrade',
                'helm.sh/hook-weight': '-5',
                'helm.sh/hook-delete-policy': 'before-hook-creation',
              } else {}),
            },
            {
              image: if helm_mode then '{{ if ((.Values.migrate_image).repository) }}{{ .Values.migrate_image.repository }}{{ else }}{{ ((.Values.image).repository) | default !"%s!" }}{{ end }}' % migrate_image else migrate_image,
              tag: if helm_mode then '{{ if ((.Values.migrate_image).tag) }}{{ .Values.migrate_image.tag }}{{ else }}{{ ((.Values.image).tag) | default !"%s!" }}{{ end }}' % migrate_tag else migrate_tag,
              command: if std.objectHas(info, 'migrate_command') && info.migrate_command != null then info.migrate_command else ['/bin/sh'],
              serviceAccount: 'init-db-%s' % [fixed.name],
              imagePullSecrets: imagePullSecrets,
              args: if std.objectHas(info, 'migrate_args') && info.migrate_args != null then info.migrate_args else [
                '-c',
                'export REAL_DSN=`echo $%s | sed -e "s/REPLACEME/${DATABASE_PASSWORD}/g"`; /usr/bin/migrate -source file:///migrations -database $REAL_DSN up' % [info.dsn.name],
              ],
              volumes: (if std.objectHas(info, 'volumes') then info.volumes(metadata) else []),
              initContainers: [
                {
                  name: 'initdb',
                  image: 'quay.io/peridot/initdb:v0.1.6',
                  command: ['/bin/sh'],
                  args: ['-c', '/bundle/initdb*'],
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
                      value: db.dsn('postgres', true),
                    },
                    {
                      name: 'INITDB_SKIP',
                      value: if helm_mode then '!!{{ if .Values.databaseUrl }}true{{ else }}false{{ end }}!!' else 'false',
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
                'linkerd.io/inject': 'disabled',
              },
            }
          ) else {},
        ]),
      [deployment]: manifestYamlStream([
        kubernetes.define_deployment(
          metadata {
            annotations: if helm_mode then {
              'resf.org/calculated-image': info.image,
              'resf.org/calculated-tag': info.tag,
            } else null
          },
          {
            replicas: if helm_mode then '{{ .Values.replicas | default !"1!" }}' else (if std.objectHas(info, 'replicas') then info.replicas else 1),
            image: image,
            tag: tag,
            command: if std.objectHas(info, 'command') then [info.command] else null,
            fsGroup: if std.objectHas(info, 'fsGroup') then info.fsGroup else null,
            fsUser: if std.objectHas(info, 'fsUser') then info.fsUser else null,
            imagePullSecrets: imagePullSecrets,
            annotations: (if std.objectHas(info, 'annotations') then info.annotations else {}) + (if disableMetrics then {} else {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '7332',
            }),
            volumes: (if std.objectHas(info, 'volumes') then info.volumes(metadata) else []),
            ports: [utils.filterObjectFields(port, ['expose']) for port in ports],
            health: if std.objectHas(info, 'health') then info.health,
            env: env + (if dbname != '' && info.backend then ([dbPassEnv]) else []) + [
              {
                name: 'SELF_IDENTITY',
                value: 'spiffe://cluster.local/ns/%s/sa/%s' % [fixed.namespace, fixed.name],
              },
            ] + [
              if std.objectHas(srv, 'expose') && srv.expose then (if helm_mode then {
                name: '%s_PUBLIC_URL' % [std.asciiUpper(std.strReplace(std.strReplace(srv.name, stage, ''), '-', '_'))],
                value: 'https://{{ .Values.%s.ingressHost }}!!' % [srv.name],
              } else {
                name: '%s_PUBLIC_URL' % [std.asciiUpper(std.strReplace(std.strReplace(srv.name, stage, ''), '-', '_'))],
                value: 'https://%s' % mappings.get(srv.name, user),
              }) else null,
            for srv in services],
            limits: if std.objectHas(info, 'limits') then info.limits,
            requests: if std.objectHas(info, 'requests') then info.requests,
            args: if std.objectHas(info, 'args') then info.args else [],
            node_pool_request: if std.objectHas(info, 'node_pool_request') then info.node_pool_request else null,
            serviceAccount: sa_name,
          },
        ),
      ]),
      [svcVsDr]:
        manifestYamlStream(
          ([kubernetes.define_service(
            metadata {
              name: srv.name,
              annotations: {
                'konghq.com/protocol': std.strReplace(std.strReplace(std.strReplace(srv.name, metadata.name, ''), stage, ''), '-', ''),
              }
            },
            srv.port,
            srv.port,
            portName=srv.portName,
            selector=metadata.name,
            env=mappings.get_env_from_svc(srv.name)
          ) for srv in services]) +
          (if istio_mode then [] else [if std.objectHas(srv, 'expose') && srv.expose then kubernetes.define_ingress(
            metadata {
              name: srv.name,
              annotations: ingress_annotations + {
                'kubernetes.io/ingress.class': if helm_mode then '{{ .Values.ingressClass | default !"!" }}' else 'kong',
                // Secure only by default
                // This produces https, grpcs, etc.
                // todo(mustafa): check if we need to add an exemption to a protocol (TCP comes to mind)
                'konghq.com/protocols': (if helm_mode then '{{ .Values.kongProtocols | default !"%ss!" }}' else '%ss') % std.strReplace(std.strReplace(std.strReplace(srv.name, metadata.name, ''), stage, ''), '-', ''),
              }
            },
            host=if helm_mode then '{{ .Values.%s.ingressHost }}' % srv.name else mappings.get(srv.name, user),
            port=srv.port,
            srvName=srv.name + '-service',
          ) else null for srv in services]) +
          (if !istio_mode then [] else [kubernetes.define_virtual_service(metadata { name: srv.name + '-internal' }, {
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
          },) for srv in services]) +
          (if !istio_mode then [] else [if std.objectHas(srv, 'expose') && srv.expose then kubernetes.define_virtual_service(
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
          ) else null for srv in services]) +
          (if !istio_mode then [] else [{
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
          } for srv in services]) +
          (if !istio_mode then [] else [{
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
          } for srv in services]) +
          (if !istio_mode then [] else [kubernetes.define_destination_rule(metadata { name: srv.name }, {
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
          },) for srv in services])
        ),
      [if std.objectHas(info, 'custom_job_items') then custom else null]:
        manifestYamlStream(if std.objectHas(info, 'custom_job_items') then info.custom_job_items(metadata, extra_info) else [{}]),
    },
}
