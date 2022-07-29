local stage = std.extVar('stage');
local ociRegistry = std.extVar('oci_registry');
local ociRegistryDocker = std.extVar('oci_registry_docker');
local localEnvironment = std.extVar('local_environment');
local localImage = if localEnvironment == "1" then true else false;

{
  local_image: localImage,
  docker_hub_image(name): "%s/%s" % [ociRegistryDocker, name],
  helm_mode: false,
}
