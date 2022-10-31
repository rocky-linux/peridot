local stage = std.extVar('stage');
local ociRegistry = std.extVar('oci_registry');
local ociRegistryDocker = std.extVar('oci_registry_docker');
local localEnvironment = std.extVar('local_environment');
local origUser = std.extVar('user');
local domainUser = std.extVar('domain_user');

local localImage = if localEnvironment == "1" then true else false;
local helm_mode = std.extVar('helm_mode') == 'true';
local stage = if helm_mode then '-{{ template !"resf.stage!" . }}' else std.extVar('stage');
local user = if domainUser != 'user-orig' then domainUser else origUser;
local stage_no_dash = std.strReplace(stage, '-', '');

{
  local_image: if helm_mode then false else localImage,
  docker_hub_image(name): "%s/%s" % [ociRegistryDocker, name],
  helm_mode: helm_mode,
  stage: stage,
  user: user,
  stage_no_dash: stage_no_dash,
}
