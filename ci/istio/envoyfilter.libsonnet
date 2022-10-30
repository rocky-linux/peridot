local base = import 'ci/istio/base.libsonnet';

{
  FILTER_TYPE_CUSTOM: 'custom',
  FILTER_TYPE_REDIRECT: 'redirect',
  MATCH_TYPE_RESFDEPLOY: 'RESFDEPLOY',
  MATCH_TYPE_ALL: 'all',

  envoy_filter(info)::
    local filterType = info.type;
    local matchType = if filterType != $.FILTER_TYPE_CUSTOM then info.matchType;

    local envoyFilterInfo = {
      kind: base.KIND_ENVOYFILTER,
      metadata: {
        name: info.name,
        namespace: 'istio-system',
      },
      spec: if filterType == $.FILTER_TYPE_CUSTOM then info.filter else {
        workloadSelector: if matchType == $.MATCH_TYPE_ALL then {}
        else if matchType == $.MATCH_TYPE_RESFDEPLOY then {
        }
        else non_existing_value,
      },
    };

    base(pubSubTopicInfo),
}
