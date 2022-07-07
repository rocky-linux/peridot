local utils = import 'ci/utils.jsonnet';

{
  hostport: if utils.local_image then 'temporal-frontend.default.svc.cluster.local:7233' else 'workflow-temporal-frontend.workflow.svc.cluster.local:7233',
  kube_env(prefix): {
    name: '%s_TEMPORAL_HOSTPORT' % prefix,
    value: $.hostport,
  }
}
