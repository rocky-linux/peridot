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

// V1ListBuildsResponse struct for V1ListBuildsResponse
type V1ListBuildsResponse struct {
	Builds *[]V1Build `json:"builds,omitempty"`
	Total *string `json:"total,omitempty"`
	Size *int32 `json:"size,omitempty"`
	Page *int32 `json:"page,omitempty"`
}

// NewV1ListBuildsResponse instantiates a new V1ListBuildsResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1ListBuildsResponse() *V1ListBuildsResponse {
	this := V1ListBuildsResponse{}
	return &this
}

// NewV1ListBuildsResponseWithDefaults instantiates a new V1ListBuildsResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1ListBuildsResponseWithDefaults() *V1ListBuildsResponse {
	this := V1ListBuildsResponse{}
	return &this
}

// GetBuilds returns the Builds field value if set, zero value otherwise.
func (o *V1ListBuildsResponse) GetBuilds() []V1Build {
	if o == nil || o.Builds == nil {
		var ret []V1Build
		return ret
	}
	return *o.Builds
}

// GetBuildsOk returns a tuple with the Builds field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBuildsResponse) GetBuildsOk() (*[]V1Build, bool) {
	if o == nil || o.Builds == nil {
		return nil, false
	}
	return o.Builds, true
}

// HasBuilds returns a boolean if a field has been set.
func (o *V1ListBuildsResponse) HasBuilds() bool {
	if o != nil && o.Builds != nil {
		return true
	}

	return false
}

// SetBuilds gets a reference to the given []V1Build and assigns it to the Builds field.
func (o *V1ListBuildsResponse) SetBuilds(v []V1Build) {
	o.Builds = &v
}

// GetTotal returns the Total field value if set, zero value otherwise.
func (o *V1ListBuildsResponse) GetTotal() string {
	if o == nil || o.Total == nil {
		var ret string
		return ret
	}
	return *o.Total
}

// GetTotalOk returns a tuple with the Total field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBuildsResponse) GetTotalOk() (*string, bool) {
	if o == nil || o.Total == nil {
		return nil, false
	}
	return o.Total, true
}

// HasTotal returns a boolean if a field has been set.
func (o *V1ListBuildsResponse) HasTotal() bool {
	if o != nil && o.Total != nil {
		return true
	}

	return false
}

// SetTotal gets a reference to the given string and assigns it to the Total field.
func (o *V1ListBuildsResponse) SetTotal(v string) {
	o.Total = &v
}

// GetSize returns the Size field value if set, zero value otherwise.
func (o *V1ListBuildsResponse) GetSize() int32 {
	if o == nil || o.Size == nil {
		var ret int32
		return ret
	}
	return *o.Size
}

// GetSizeOk returns a tuple with the Size field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBuildsResponse) GetSizeOk() (*int32, bool) {
	if o == nil || o.Size == nil {
		return nil, false
	}
	return o.Size, true
}

// HasSize returns a boolean if a field has been set.
func (o *V1ListBuildsResponse) HasSize() bool {
	if o != nil && o.Size != nil {
		return true
	}

	return false
}

// SetSize gets a reference to the given int32 and assigns it to the Size field.
func (o *V1ListBuildsResponse) SetSize(v int32) {
	o.Size = &v
}

// GetPage returns the Page field value if set, zero value otherwise.
func (o *V1ListBuildsResponse) GetPage() int32 {
	if o == nil || o.Page == nil {
		var ret int32
		return ret
	}
	return *o.Page
}

// GetPageOk returns a tuple with the Page field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1ListBuildsResponse) GetPageOk() (*int32, bool) {
	if o == nil || o.Page == nil {
		return nil, false
	}
	return o.Page, true
}

// HasPage returns a boolean if a field has been set.
func (o *V1ListBuildsResponse) HasPage() bool {
	if o != nil && o.Page != nil {
		return true
	}

	return false
}

// SetPage gets a reference to the given int32 and assigns it to the Page field.
func (o *V1ListBuildsResponse) SetPage(v int32) {
	o.Page = &v
}

func (o V1ListBuildsResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Builds != nil {
		toSerialize["builds"] = o.Builds
	}
	if o.Total != nil {
		toSerialize["total"] = o.Total
	}
	if o.Size != nil {
		toSerialize["size"] = o.Size
	}
	if o.Page != nil {
		toSerialize["page"] = o.Page
	}
	return json.Marshal(toSerialize)
}

type NullableV1ListBuildsResponse struct {
	value *V1ListBuildsResponse
	isSet bool
}

func (v NullableV1ListBuildsResponse) Get() *V1ListBuildsResponse {
	return v.value
}

func (v *NullableV1ListBuildsResponse) Set(val *V1ListBuildsResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableV1ListBuildsResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableV1ListBuildsResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1ListBuildsResponse(val *V1ListBuildsResponse) *NullableV1ListBuildsResponse {
	return &NullableV1ListBuildsResponse{value: val, isSet: true}
}

func (v NullableV1ListBuildsResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1ListBuildsResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


