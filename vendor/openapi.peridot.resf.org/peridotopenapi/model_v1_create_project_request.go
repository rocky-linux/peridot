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
)

// V1CreateProjectRequest struct for V1CreateProjectRequest
type V1CreateProjectRequest struct {
	Project *V1Project `json:"project,omitempty"`
}

// NewV1CreateProjectRequest instantiates a new V1CreateProjectRequest object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1CreateProjectRequest() *V1CreateProjectRequest {
	this := V1CreateProjectRequest{}
	return &this
}

// NewV1CreateProjectRequestWithDefaults instantiates a new V1CreateProjectRequest object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1CreateProjectRequestWithDefaults() *V1CreateProjectRequest {
	this := V1CreateProjectRequest{}
	return &this
}

// GetProject returns the Project field value if set, zero value otherwise.
func (o *V1CreateProjectRequest) GetProject() V1Project {
	if o == nil || o.Project == nil {
		var ret V1Project
		return ret
	}
	return *o.Project
}

// GetProjectOk returns a tuple with the Project field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1CreateProjectRequest) GetProjectOk() (*V1Project, bool) {
	if o == nil || o.Project == nil {
		return nil, false
	}
	return o.Project, true
}

// HasProject returns a boolean if a field has been set.
func (o *V1CreateProjectRequest) HasProject() bool {
	if o != nil && o.Project != nil {
		return true
	}

	return false
}

// SetProject gets a reference to the given V1Project and assigns it to the Project field.
func (o *V1CreateProjectRequest) SetProject(v V1Project) {
	o.Project = &v
}

func (o V1CreateProjectRequest) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Project != nil {
		toSerialize["project"] = o.Project
	}
	return json.Marshal(toSerialize)
}

type NullableV1CreateProjectRequest struct {
	value *V1CreateProjectRequest
	isSet bool
}

func (v NullableV1CreateProjectRequest) Get() *V1CreateProjectRequest {
	return v.value
}

func (v *NullableV1CreateProjectRequest) Set(val *V1CreateProjectRequest) {
	v.value = val
	v.isSet = true
}

func (v NullableV1CreateProjectRequest) IsSet() bool {
	return v.isSet
}

func (v *NullableV1CreateProjectRequest) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1CreateProjectRequest(val *V1CreateProjectRequest) *NullableV1CreateProjectRequest {
	return &NullableV1CreateProjectRequest{value: val, isSet: true}
}

func (v NullableV1CreateProjectRequest) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1CreateProjectRequest) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


