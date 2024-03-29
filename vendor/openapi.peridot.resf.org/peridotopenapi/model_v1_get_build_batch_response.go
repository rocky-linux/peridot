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

// V1GetBuildBatchResponse struct for V1GetBuildBatchResponse
type V1GetBuildBatchResponse struct {
	Builds *[]V1Build `json:"builds,omitempty"`
	Pending *int32 `json:"pending,omitempty"`
	Running *int32 `json:"running,omitempty"`
	Succeeded *int32 `json:"succeeded,omitempty"`
	Failed *int32 `json:"failed,omitempty"`
	Canceled *int32 `json:"canceled,omitempty"`
	Total *string `json:"total,omitempty"`
	Size *int32 `json:"size,omitempty"`
	Page *int32 `json:"page,omitempty"`
}

// NewV1GetBuildBatchResponse instantiates a new V1GetBuildBatchResponse object
// This constructor will assign default values to properties that have it defined,
// and makes sure properties required by API are set, but the set of arguments
// will change when the set of required properties is changed
func NewV1GetBuildBatchResponse() *V1GetBuildBatchResponse {
	this := V1GetBuildBatchResponse{}
	return &this
}

// NewV1GetBuildBatchResponseWithDefaults instantiates a new V1GetBuildBatchResponse object
// This constructor will only assign default values to properties that have it defined,
// but it doesn't guarantee that properties required by API are set
func NewV1GetBuildBatchResponseWithDefaults() *V1GetBuildBatchResponse {
	this := V1GetBuildBatchResponse{}
	return &this
}

// GetBuilds returns the Builds field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetBuilds() []V1Build {
	if o == nil || o.Builds == nil {
		var ret []V1Build
		return ret
	}
	return *o.Builds
}

// GetBuildsOk returns a tuple with the Builds field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetBuildsOk() (*[]V1Build, bool) {
	if o == nil || o.Builds == nil {
		return nil, false
	}
	return o.Builds, true
}

// HasBuilds returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasBuilds() bool {
	if o != nil && o.Builds != nil {
		return true
	}

	return false
}

// SetBuilds gets a reference to the given []V1Build and assigns it to the Builds field.
func (o *V1GetBuildBatchResponse) SetBuilds(v []V1Build) {
	o.Builds = &v
}

// GetPending returns the Pending field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetPending() int32 {
	if o == nil || o.Pending == nil {
		var ret int32
		return ret
	}
	return *o.Pending
}

// GetPendingOk returns a tuple with the Pending field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetPendingOk() (*int32, bool) {
	if o == nil || o.Pending == nil {
		return nil, false
	}
	return o.Pending, true
}

// HasPending returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasPending() bool {
	if o != nil && o.Pending != nil {
		return true
	}

	return false
}

// SetPending gets a reference to the given int32 and assigns it to the Pending field.
func (o *V1GetBuildBatchResponse) SetPending(v int32) {
	o.Pending = &v
}

// GetRunning returns the Running field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetRunning() int32 {
	if o == nil || o.Running == nil {
		var ret int32
		return ret
	}
	return *o.Running
}

// GetRunningOk returns a tuple with the Running field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetRunningOk() (*int32, bool) {
	if o == nil || o.Running == nil {
		return nil, false
	}
	return o.Running, true
}

// HasRunning returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasRunning() bool {
	if o != nil && o.Running != nil {
		return true
	}

	return false
}

// SetRunning gets a reference to the given int32 and assigns it to the Running field.
func (o *V1GetBuildBatchResponse) SetRunning(v int32) {
	o.Running = &v
}

// GetSucceeded returns the Succeeded field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetSucceeded() int32 {
	if o == nil || o.Succeeded == nil {
		var ret int32
		return ret
	}
	return *o.Succeeded
}

// GetSucceededOk returns a tuple with the Succeeded field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetSucceededOk() (*int32, bool) {
	if o == nil || o.Succeeded == nil {
		return nil, false
	}
	return o.Succeeded, true
}

// HasSucceeded returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasSucceeded() bool {
	if o != nil && o.Succeeded != nil {
		return true
	}

	return false
}

// SetSucceeded gets a reference to the given int32 and assigns it to the Succeeded field.
func (o *V1GetBuildBatchResponse) SetSucceeded(v int32) {
	o.Succeeded = &v
}

// GetFailed returns the Failed field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetFailed() int32 {
	if o == nil || o.Failed == nil {
		var ret int32
		return ret
	}
	return *o.Failed
}

// GetFailedOk returns a tuple with the Failed field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetFailedOk() (*int32, bool) {
	if o == nil || o.Failed == nil {
		return nil, false
	}
	return o.Failed, true
}

// HasFailed returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasFailed() bool {
	if o != nil && o.Failed != nil {
		return true
	}

	return false
}

// SetFailed gets a reference to the given int32 and assigns it to the Failed field.
func (o *V1GetBuildBatchResponse) SetFailed(v int32) {
	o.Failed = &v
}

// GetCanceled returns the Canceled field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetCanceled() int32 {
	if o == nil || o.Canceled == nil {
		var ret int32
		return ret
	}
	return *o.Canceled
}

// GetCanceledOk returns a tuple with the Canceled field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetCanceledOk() (*int32, bool) {
	if o == nil || o.Canceled == nil {
		return nil, false
	}
	return o.Canceled, true
}

// HasCanceled returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasCanceled() bool {
	if o != nil && o.Canceled != nil {
		return true
	}

	return false
}

// SetCanceled gets a reference to the given int32 and assigns it to the Canceled field.
func (o *V1GetBuildBatchResponse) SetCanceled(v int32) {
	o.Canceled = &v
}

// GetTotal returns the Total field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetTotal() string {
	if o == nil || o.Total == nil {
		var ret string
		return ret
	}
	return *o.Total
}

// GetTotalOk returns a tuple with the Total field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetTotalOk() (*string, bool) {
	if o == nil || o.Total == nil {
		return nil, false
	}
	return o.Total, true
}

// HasTotal returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasTotal() bool {
	if o != nil && o.Total != nil {
		return true
	}

	return false
}

// SetTotal gets a reference to the given string and assigns it to the Total field.
func (o *V1GetBuildBatchResponse) SetTotal(v string) {
	o.Total = &v
}

// GetSize returns the Size field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetSize() int32 {
	if o == nil || o.Size == nil {
		var ret int32
		return ret
	}
	return *o.Size
}

// GetSizeOk returns a tuple with the Size field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetSizeOk() (*int32, bool) {
	if o == nil || o.Size == nil {
		return nil, false
	}
	return o.Size, true
}

// HasSize returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasSize() bool {
	if o != nil && o.Size != nil {
		return true
	}

	return false
}

// SetSize gets a reference to the given int32 and assigns it to the Size field.
func (o *V1GetBuildBatchResponse) SetSize(v int32) {
	o.Size = &v
}

// GetPage returns the Page field value if set, zero value otherwise.
func (o *V1GetBuildBatchResponse) GetPage() int32 {
	if o == nil || o.Page == nil {
		var ret int32
		return ret
	}
	return *o.Page
}

// GetPageOk returns a tuple with the Page field value if set, nil otherwise
// and a boolean to check if the value has been set.
func (o *V1GetBuildBatchResponse) GetPageOk() (*int32, bool) {
	if o == nil || o.Page == nil {
		return nil, false
	}
	return o.Page, true
}

// HasPage returns a boolean if a field has been set.
func (o *V1GetBuildBatchResponse) HasPage() bool {
	if o != nil && o.Page != nil {
		return true
	}

	return false
}

// SetPage gets a reference to the given int32 and assigns it to the Page field.
func (o *V1GetBuildBatchResponse) SetPage(v int32) {
	o.Page = &v
}

func (o V1GetBuildBatchResponse) MarshalJSON() ([]byte, error) {
	toSerialize := map[string]interface{}{}
	if o.Builds != nil {
		toSerialize["builds"] = o.Builds
	}
	if o.Pending != nil {
		toSerialize["pending"] = o.Pending
	}
	if o.Running != nil {
		toSerialize["running"] = o.Running
	}
	if o.Succeeded != nil {
		toSerialize["succeeded"] = o.Succeeded
	}
	if o.Failed != nil {
		toSerialize["failed"] = o.Failed
	}
	if o.Canceled != nil {
		toSerialize["canceled"] = o.Canceled
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

type NullableV1GetBuildBatchResponse struct {
	value *V1GetBuildBatchResponse
	isSet bool
}

func (v NullableV1GetBuildBatchResponse) Get() *V1GetBuildBatchResponse {
	return v.value
}

func (v *NullableV1GetBuildBatchResponse) Set(val *V1GetBuildBatchResponse) {
	v.value = val
	v.isSet = true
}

func (v NullableV1GetBuildBatchResponse) IsSet() bool {
	return v.isSet
}

func (v *NullableV1GetBuildBatchResponse) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1GetBuildBatchResponse(val *V1GetBuildBatchResponse) *NullableV1GetBuildBatchResponse {
	return &NullableV1GetBuildBatchResponse{value: val, isSet: true}
}

func (v NullableV1GetBuildBatchResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1GetBuildBatchResponse) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}


