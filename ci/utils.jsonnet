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

  // Function to filter an object by excluding specified fields.
  // Parameters:
  // - inputObject: The object to be filtered.
  // - fieldsToIgnore: List of fields to be ignored from the input object.
  filterObjectFields(inputObject, fieldsToIgnore)::
    // Iterating over the fields in the input object and creating a new object
    // without the fields specified in `fieldsToIgnore`.
    std.foldl(function(filteredObject, currentField)
      // If current field is in `fieldsToIgnore`, return the filtered object as is.
      // Otherwise, add the current field to the filtered object.
      (
        if std.member(fieldsToIgnore, currentField) then
          filteredObject
        else
          filteredObject + { [currentField]: inputObject[currentField] }
      ),
      // Starting with an empty object and iterating over each field in the input object.
      std.objectFields(inputObject), {}),
}
