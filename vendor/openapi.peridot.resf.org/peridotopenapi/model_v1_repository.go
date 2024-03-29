/*
 * peridot/proto/v1/batch.proto
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: version not set
 */

// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

package peridotopenapi

import (
	"encoding/json"
	"time"
)

// V1Repository struct for V1Repository
type V1Repository struct {
	Id *string `json:"id,omitempty"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	Name *string `json:"name,omitempty"`
	ProjectId *string `json:"projectId,omitempty"`
	Packages *[]string `json:"packages,omitempty"`
	ExcludeFilter *[]string `json:"excludeFilter,omitempty"`
	// Whether an RPM from a package should be included in the repository If list contains a NA that is in exclude_list as well, then it will be excluded.
	IncludeList *[]string `json:"includeList,omitempty"`
	AdditionalMultilib *[]string `json:"additionalMultilib,omitempty"`
	ExcludeMultilibFilter *[]string `json:"excludeMultilibFilter,omitempty"`
	Multilib *[]string `json:"multilib,omitempty"`
	GlobIncludeFilter *[]string `json:"globIncludeFilter,omitempty"`
}

// NewV1Repository instantiates a new V1Repository object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1Repository() *V1Repository {
	this := V1Repository{}
	return &this
}

// NewV1RepositoryWithDefaults instantiates a new V1Repository object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1RepositoryWithDefaults() *V1Repository {
	this := V1Repository{}
	return &this
}

// GetId returns the Id field value if set, zero value otherwise.
func (o *V1Repository) GetId() string {
	if o == nil || o.Id == nil {
		var ret string
		return ret
	}
	return *o.Id
}

// GetIdOk returns a tuple with the Id field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetIdOk() (*string, bool) {
	if o == nil || o.Id == nil {
		return nil, false
	}
	return o.Id, true
}

// HasId returns a boolean if a field has been set.
func (o *V1Repository) HasId() bool {
	if o != nil && o.Id != nil {
		return true
	}

	return false
}

// SetId gets a reference to the given string and assigns it to the Id field.
func (o *V1Repository) SetId(v string) {
	o.Id = &v
}

// GetCreatedAt returns the CreatedAt field value if set, zero value otherwise.
func (o *V1Repository) GetCreatedAt() time.Time {
	if o == nil || o.CreatedAt == nil {
		var ret time.Time
		return ret
	}
	return *o.CreatedAt
}

// GetCreatedAtOk returns a tuple with the CreatedAt field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetCreatedAtOk() (*time.Time, bool) {
	if o == nil || o.CreatedAt == nil {
		return nil, false
	}
	return o.CreatedAt, true
}

// HasCreatedAt returns a boolean if a field has been set.
func (o *V1Repository) HasCreatedAt() bool {
	if o != nil && o.CreatedAt != nil {
		return true
	}

	return false
}

// SetCreatedAt gets a reference to the given time.Time and assigns it to the CreatedAt field.
func (o *V1Repository) SetCreatedAt(v time.Time) {
	o.CreatedAt = &v
}

// GetName returns the Name field value if set, zero value otherwise.
func (o *V1Repository) GetName() string {
	if o == nil || o.Name == nil {
		var ret string
		return ret
	}
	return *o.Name
}

// GetNameOk returns a tuple with the Name field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetNameOk() (*string, bool) {
	if o == nil || o.Name == nil {
		return nil, false
	}
	return o.Name, true
}

// HasName returns a boolean if a field has been set.
func (o *V1Repository) HasName() bool {
	if o != nil && o.Name != nil {
		return true
	}

	return false
}

// SetName gets a reference to the given string and assigns it to the Name field.
func (o *V1Repository) SetName(v string) {
	o.Name = &v
}

// GetProjectId returns the ProjectId field value if set, zero value otherwise.
func (o *V1Repository) GetProjectId() string {
	if o == nil || o.ProjectId == nil {
		var ret string
		return ret
	}
	return *o.ProjectId
}

// GetProjectIdOk returns a tuple with the ProjectId field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetProjectIdOk() (*string, bool) {
	if o == nil || o.ProjectId == nil {
		return nil, false
	}
	return o.ProjectId, true
}

// HasProjectId returns a boolean if a field has been set.
func (o *V1Repository) HasProjectId() bool {
	if o != nil && o.ProjectId != nil {
		return true
	}

	return false
}

// SetProjectId gets a reference to the given string and assigns it to the ProjectId field.
func (o *V1Repository) SetProjectId(v string) {
	o.ProjectId = &v
}

// GetPackages returns the Packages field value if set, zero value otherwise.
func (o *V1Repository) GetPackages() []string {
	if o == nil || o.Packages == nil {
		var ret []string
		return ret
	}
	return *o.Packages
}

// GetPackagesOk returns a tuple with the Packages field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetPackagesOk() (*[]string, bool) {
	if o == nil || o.Packages == nil {
		return nil, false
	}
	return o.Packages, true
}

// HasPackages returns a boolean if a field has been set.
func (o *V1Repository) HasPackages() bool {
	if o != nil && o.Packages != nil {
		return true
	}

	return false
}

// SetPackages gets a reference to the given []string and assigns it to the Packages field.
func (o *V1Repository) SetPackages(v []string) {
	o.Packages = &v
}

// GetExcludeFilter returns the ExcludeFilter field value if set, zero value otherwise.
func (o *V1Repository) GetExcludeFilter() []string {
	if o == nil || o.ExcludeFilter == nil {
		var ret []string
		return ret
	}
	return *o.ExcludeFilter
}

// GetExcludeFilterOk returns a tuple with the ExcludeFilter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetExcludeFilterOk() (*[]string, bool) {
	if o == nil || o.ExcludeFilter == nil {
		return nil, false
	}
	return o.ExcludeFilter, true
}

// HasExcludeFilter returns a boolean if a field has been set.
func (o *V1Repository) HasExcludeFilter() bool {
	if o != nil && o.ExcludeFilter != nil {
		return true
	}

	return false
}

// SetExcludeFilter gets a reference to the given []string and assigns it to the ExcludeFilter field.
func (o *V1Repository) SetExcludeFilter(v []string) {
	o.ExcludeFilter = &v
}

// GetIncludeList returns the IncludeList field value if set, zero value otherwise.
func (o *V1Repository) GetIncludeList() []string {
	if o == nil || o.IncludeList == nil {
		var ret []string
		return ret
	}
	return *o.IncludeList
}

// GetIncludeListOk returns a tuple with the IncludeList field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetIncludeListOk() (*[]string, bool) {
	if o == nil || o.IncludeList == nil {
		return nil, false
	}
	return o.IncludeList, true
}

// HasIncludeList returns a boolean if a field has been set.
func (o *V1Repository) HasIncludeList() bool {
	if o != nil && o.IncludeList != nil {
		return true
	}

	return false
}

// SetIncludeList gets a reference to the given []string and assigns it to the IncludeList field.
func (o *V1Repository) SetIncludeList(v []string) {
	o.IncludeList = &v
}

// GetAdditionalMultilib returns the AdditionalMultilib field value if set, zero value otherwise.
func (o *V1Repository) GetAdditionalMultilib() []string {
	if o == nil || o.AdditionalMultilib == nil {
		var ret []string
		return ret
	}
	return *o.AdditionalMultilib
}

// GetAdditionalMultilibOk returns a tuple with the AdditionalMultilib field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetAdditionalMultilibOk() (*[]string, bool) {
	if o == nil || o.AdditionalMultilib == nil {
		return nil, false
	}
	return o.AdditionalMultilib, true
}

// HasAdditionalMultilib returns a boolean if a field has been set.
func (o *V1Repository) HasAdditionalMultilib() bool {
	if o != nil && o.AdditionalMultilib != nil {
		return true
	}

	return false
}

// SetAdditionalMultilib gets a reference to the given []string and assigns it to the AdditionalMultilib field.
func (o *V1Repository) SetAdditionalMultilib(v []string) {
	o.AdditionalMultilib = &v
}

// GetExcludeMultilibFilter returns the ExcludeMultilibFilter field value if set, zero value otherwise.
func (o *V1Repository) GetExcludeMultilibFilter() []string {
	if o == nil || o.ExcludeMultilibFilter == nil {
		var ret []string
		return ret
	}
	return *o.ExcludeMultilibFilter
}

// GetExcludeMultilibFilterOk returns a tuple with the ExcludeMultilibFilter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetExcludeMultilibFilterOk() (*[]string, bool) {
	if o == nil || o.ExcludeMultilibFilter == nil {
		return nil, false
	}
	return o.ExcludeMultilibFilter, true
}

// HasExcludeMultilibFilter returns a boolean if a field has been set.
func (o *V1Repository) HasExcludeMultilibFilter() bool {
	if o != nil && o.ExcludeMultilibFilter != nil {
		return true
	}

	return false
}

// SetExcludeMultilibFilter gets a reference to the given []string and assigns it to the ExcludeMultilibFilter field.
func (o *V1Repository) SetExcludeMultilibFilter(v []string) {
	o.ExcludeMultilibFilter = &v
}

// GetMultilib returns the Multilib field value if set, zero value otherwise.
func (o *V1Repository) GetMultilib() []string {
	if o == nil || o.Multilib == nil {
		var ret []string
		return ret
	}
	return *o.Multilib
}

// GetMultilibOk returns a tuple with the Multilib field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetMultilibOk() (*[]string, bool) {
	if o == nil || o.Multilib == nil {
		return nil, false
	}
	return o.Multilib, true
}

// HasMultilib returns a boolean if a field has been set.
func (o *V1Repository) HasMultilib() bool {
	if o != nil && o.Multilib != nil {
		return true
	}

	return false
}

// SetMultilib gets a reference to the given []string and assigns it to the Multilib field.
func (o *V1Repository) SetMultilib(v []string) {
	o.Multilib = &v
}

// GetGlobIncludeFilter returns the GlobIncludeFilter field value if set, zero value otherwise.
func (o *V1Repository) GetGlobIncludeFilter() []string {
	if o == nil || o.GlobIncludeFilter == nil {
		var ret []string
		return ret
	}
	return *o.GlobIncludeFilter
}

// GetGlobIncludeFilterOk returns a tuple with the GlobIncludeFilter field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1Repository) GetGlobIncludeFilterOk() (*[]string, bool) {
	if o == nil || o.GlobIncludeFilter == nil {
		return nil, false
	}
	return o.GlobIncludeFilter, true
}

// HasGlobIncludeFilter returns a boolean if a field has been set.
func (o *V1Repository) HasGlobIncludeFilter() bool {
	if o != nil && o.GlobIncludeFilter != nil {
		return true
	}

	return false
}

// SetGlobIncludeFilter gets a reference to the given []string and assigns it to the GlobIncludeFilter field.
func (o *V1Repository) SetGlobIncludeFilter(v []string) {
	o.GlobIncludeFilter = &v
}

func (o V1Repository) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Id != nil {
		toSerialize["id"] = o.Id
	}
	if o.CreatedAt != nil {
		toSerialize["createdAt"] = o.CreatedAt
	}
	if o.Name != nil {
		toSerialize["name"] = o.Name
	}
	if o.ProjectId != nil {
		toSerialize["projectId"] = o.ProjectId
	}
	if o.Packages != nil {
		toSerialize["packages"] = o.Packages
	}
	if o.ExcludeFilter != nil {
		toSerialize["excludeFilter"] = o.ExcludeFilter
	}
	if o.IncludeList != nil {
		toSerialize["includeList"] = o.IncludeList
	}
	if o.AdditionalMultilib != nil {
		toSerialize["additionalMultilib"] = o.AdditionalMultilib
	}
	if o.ExcludeMultilibFilter != nil {
		toSerialize["excludeMultilibFilter"] = o.ExcludeMultilibFilter
	}
	if o.Multilib != nil {
		toSerialize["multilib"] = o.Multilib
	}
	if o.GlobIncludeFilter != nil {
		toSerialize["globIncludeFilter"] = o.GlobIncludeFilter
	}
	return json.Marshal(toSerialize)
}

type NullableV1Repository struct {
	value *V1Repository
	isSet bool
}

func (v NullableV1Repository) Get() *V1Repository {
	return v.value
}

func (v *NullableV1Repository) Set(val *V1Repository) {
	v.value = val
	v.isSet = true
}

func (v NullableV1Repository) IsSet() bool {
	return v.isSet
}

func (v *NullableV1Repository) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1Repository(val *V1Repository) *NullableV1Repository {
	return &NullableV1Repository{value: val, isSet: true}
}

func (v NullableV1Repository) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1Repository) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


