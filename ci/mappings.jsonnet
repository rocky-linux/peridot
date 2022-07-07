local utils = import 'ci/utils.jsonnet';
local sm = import 'ci/service_mappings.jsonnet';

sm.service_mappings + {
  get_env_from_svc(svc_name)::
    local spl = std.split(svc_name, '-');
    spl[std.length(spl) - 1],

  get_obj_by_svc(svc_name)::
    local env = $.get_env_from_svc(svc_name);
    local key = std.strReplace(svc_name, '-%s' % env, '');

    if std.objectHas($, key) then $[key] else null,

  is_external(svc_name)::
    local mapping = $.get_obj_by_svc(svc_name);
    if utils.local_image then true else if mapping != null then std.objectHas(mapping, 'external') && mapping.external else false,

  should_expose_all(svc_name)::
    local mapping = $.get_obj_by_svc(svc_name);
    if mapping != null then std.objectHas(mapping, 'expose_all_envs') && mapping.expose_all_envs else false,

  get(svc_name, user='')::
    local env = $.get_env_from_svc(svc_name);
    local mapping = $.get_obj_by_svc(svc_name);
    local prefix = if mapping.id == '*' then (if env != 'prod' then mapping.devId else mapping.id) else mapping.id;
    local devId = if std.objectHas(mapping, 'devId') then '%s-' % mapping.devId else '';

    local prod_def_domain = if utils.local_image then sm.local_domain else (if std.objectHas(mapping, 'domain') then mapping.domain else sm.default_domain);

    '%s%s' % [prefix, if env == 'prod' then prod_def_domain else '-%s.internal%s' % [if env == 'dev' && !utils.local_image then '%s%s' % [devId, user] else env, prod_def_domain]],
}
