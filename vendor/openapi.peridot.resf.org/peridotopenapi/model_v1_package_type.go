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
	"fmt"
)

// V1PackageType - PACKAGE_TYPE_DEFAULT: Unknown value. Should never be used  - PACKAGE_TYPE_NORMAL: Normal packages from downstream dist-git The repos are imported as-is This will never be used as PACKAGE_TYPE_NORMAL_FORK accomplishes the same task without duplicate work  - PACKAGE_TYPE_NORMAL_FORK: Normal packages from upstream dist-git The repos are first imported into target dist-git using srpmproc (with eventual patches) and then imported as-is into Peridot  - PACKAGE_TYPE_NORMAL_SRC: Source packages from downstream src-git The sources are packaged into tarballs and uploaded into lookaside, and the repo with sources removed is then pushed into dist-git with a following metadata file. This package type enables an automatic src-git packaging workflow, but a manual workflow may be adapted as well with manual packaging. The package should then be set to PACKAGE_TYPE_NORMAL if manual packaging is desired.  - PACKAGE_TYPE_MODULE_FORK: todo(mustafa): Document rest PACKAGE_TYPE_MODULE = 4; PACKAGE_TYPE_MODULE_COMPONENT = 5;  - PACKAGE_TYPE_NORMAL_FORK_MODULE: A package may be both a normally forked package and a module So we need to differentiate between the two  - PACKAGE_TYPE_NORMAL_FORK_MODULE_COMPONENT: A package may also be a module component and a normal package So we need to differentiate between the two  - PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT: A package may be both a module and a module component
type V1PackageType string

// List of v1PackageType
const (
	DEFAULT V1PackageType = "PACKAGE_TYPE_DEFAULT"
	NORMAL V1PackageType = "PACKAGE_TYPE_NORMAL"
	NORMAL_FORK V1PackageType = "PACKAGE_TYPE_NORMAL_FORK"
	NORMAL_SRC V1PackageType = "PACKAGE_TYPE_NORMAL_SRC"
	MODULE_FORK V1PackageType = "PACKAGE_TYPE_MODULE_FORK"
	MODULE_FORK_COMPONENT V1PackageType = "PACKAGE_TYPE_MODULE_FORK_COMPONENT"
	NORMAL_FORK_MODULE V1PackageType = "PACKAGE_TYPE_NORMAL_FORK_MODULE"
	NORMAL_FORK_MODULE_COMPONENT V1PackageType = "PACKAGE_TYPE_NORMAL_FORK_MODULE_COMPONENT"
	MODULE_FORK_MODULE_COMPONENT V1PackageType = "PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT"
)

func (v *V1PackageType) UnmarshalJSON(src []byte) error {
	var value string
	err := json.Unmarshal(src, &value)
	if err != nil {
		return err
	}
	enumTypeValue := V1PackageType(value)
	for _, existing := range []V1PackageType{ "PACKAGE_TYPE_DEFAULT", "PACKAGE_TYPE_NORMAL", "PACKAGE_TYPE_NORMAL_FORK", "PACKAGE_TYPE_NORMAL_SRC", "PACKAGE_TYPE_MODULE_FORK", "PACKAGE_TYPE_MODULE_FORK_COMPONENT", "PACKAGE_TYPE_NORMAL_FORK_MODULE", "PACKAGE_TYPE_NORMAL_FORK_MODULE_COMPONENT", "PACKAGE_TYPE_MODULE_FORK_MODULE_COMPONENT",   } {
		if existing == enumTypeValue {
			*v = enumTypeValue
			return nil
		}
	}

	return fmt.Errorf("%+v is not a valid V1PackageType", value)
}

// Ptr returns reference to v1PackageType value
func (v V1PackageType) Ptr() *V1PackageType {
	return &v
}

type NullableV1PackageType struct {
	value *V1PackageType
	isSet bool
}

func (v NullableV1PackageType) Get() *V1PackageType {
	return v.value
}

func (v *NullableV1PackageType) Set(val *V1PackageType) {
	v.value = val
	v.isSet = true
}

func (v NullableV1PackageType) IsSet() bool {
	return v.isSet
}

func (v *NullableV1PackageType) Unset() {
	v.value = nil
	v.isSet = false
}

func NewNullableV1PackageType(val *V1PackageType) *NullableV1PackageType {
	return &NullableV1PackageType{value: val, isSet: true}
}

func (v NullableV1PackageType) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.value)
}

func (v *NullableV1PackageType) UnmarshalJSON(src []byte) error {
	v.isSet = true
	return json.Unmarshal(src, &v.value)
}

