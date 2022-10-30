# CVEDetailed

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ThreatSeverity** | **string** |  | 
**PublicDate** | **string** |  | 
**Bugzilla** | [**CVEDetailedBugzilla**](CVEDetailedBugzilla.md) |  | 
**Cvss3** | [**CVEDetailedCvss3**](CVEDetailedCvss3.md) |  | 
**Cwe** | **string** |  | 
**Details** | **[]string** |  | 
**Acknowledgement** | **string** |  | 
**AffectedRelease** | Pointer to [**[]CVEDetailedAffectedRelease**](CVEDetailedAffectedRelease.md) |  | [optional] 
**Name** | **string** |  | 
**Csaw** | **bool** |  | 
**PackageState** | Pointer to [**[]CVEDetailedPackageState**](CVEDetailedPackageState.md) |  | [optional] 

## Methods

### NewCVEDetailed

`func NewCVEDetailed(threatSeverity string, publicDate string, bugzilla CVEDetailedBugzilla, cvss3 CVEDetailedCvss3, cwe string, details []string, acknowledgement string, name string, csaw bool, ) *CVEDetailed`

NewCVEDetailed instantiates a new CVEDetailed object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCVEDetailedWithDefaults

`func NewCVEDetailedWithDefaults() *CVEDetailed`

NewCVEDetailedWithDefaults instantiates a new CVEDetailed object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetThreatSeverity

`func (o *CVEDetailed) GetThreatSeverity() string`

GetThreatSeverity returns the ThreatSeverity field if non-nil, zero value otherwise.

### GetThreatSeverityOk

`func (o *CVEDetailed) GetThreatSeverityOk() (*string, bool)`

GetThreatSeverityOk returns a tuple with the ThreatSeverity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetThreatSeverity

`func (o *CVEDetailed) SetThreatSeverity(v string)`

SetThreatSeverity sets ThreatSeverity field to given value.


### GetPublicDate

`func (o *CVEDetailed) GetPublicDate() string`

GetPublicDate returns the PublicDate field if non-nil, zero value otherwise.

### GetPublicDateOk

`func (o *CVEDetailed) GetPublicDateOk() (*string, bool)`

GetPublicDateOk returns a tuple with the PublicDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicDate

`func (o *CVEDetailed) SetPublicDate(v string)`

SetPublicDate sets PublicDate field to given value.


### GetBugzilla

`func (o *CVEDetailed) GetBugzilla() CVEDetailedBugzilla`

GetBugzilla returns the Bugzilla field if non-nil, zero value otherwise.

### GetBugzillaOk

`func (o *CVEDetailed) GetBugzillaOk() (*CVEDetailedBugzilla, bool)`

GetBugzillaOk returns a tuple with the Bugzilla field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBugzilla

`func (o *CVEDetailed) SetBugzilla(v CVEDetailedBugzilla)`

SetBugzilla sets Bugzilla field to given value.


### GetCvss3

`func (o *CVEDetailed) GetCvss3() CVEDetailedCvss3`

GetCvss3 returns the Cvss3 field if non-nil, zero value otherwise.

### GetCvss3Ok

`func (o *CVEDetailed) GetCvss3Ok() (*CVEDetailedCvss3, bool)`

GetCvss3Ok returns a tuple with the Cvss3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCvss3

`func (o *CVEDetailed) SetCvss3(v CVEDetailedCvss3)`

SetCvss3 sets Cvss3 field to given value.


### GetCwe

`func (o *CVEDetailed) GetCwe() string`

GetCwe returns the Cwe field if non-nil, zero value otherwise.

### GetCweOk

`func (o *CVEDetailed) GetCweOk() (*string, bool)`

GetCweOk returns a tuple with the Cwe field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCwe

`func (o *CVEDetailed) SetCwe(v string)`

SetCwe sets Cwe field to given value.


### GetDetails

`func (o *CVEDetailed) GetDetails() []string`

GetDetails returns the Details field if non-nil, zero value otherwise.

### GetDetailsOk

`func (o *CVEDetailed) GetDetailsOk() (*[]string, bool)`

GetDetailsOk returns a tuple with the Details field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDetails

`func (o *CVEDetailed) SetDetails(v []string)`

SetDetails sets Details field to given value.


### GetAcknowledgement

`func (o *CVEDetailed) GetAcknowledgement() string`

GetAcknowledgement returns the Acknowledgement field if non-nil, zero value otherwise.

### GetAcknowledgementOk

`func (o *CVEDetailed) GetAcknowledgementOk() (*string, bool)`

GetAcknowledgementOk returns a tuple with the Acknowledgement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAcknowledgement

`func (o *CVEDetailed) SetAcknowledgement(v string)`

SetAcknowledgement sets Acknowledgement field to given value.


### GetAffectedRelease

`func (o *CVEDetailed) GetAffectedRelease() []CVEDetailedAffectedRelease`

GetAffectedRelease returns the AffectedRelease field if non-nil, zero value otherwise.

### GetAffectedReleaseOk

`func (o *CVEDetailed) GetAffectedReleaseOk() (*[]CVEDetailedAffectedRelease, bool)`

GetAffectedReleaseOk returns a tuple with the AffectedRelease field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffectedRelease

`func (o *CVEDetailed) SetAffectedRelease(v []CVEDetailedAffectedRelease)`

SetAffectedRelease sets AffectedRelease field to given value.

### HasAffectedRelease

`func (o *CVEDetailed) HasAffectedRelease() bool`

HasAffectedRelease returns a boolean if a field has been set.

### GetName

`func (o *CVEDetailed) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *CVEDetailed) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *CVEDetailed) SetName(v string)`

SetName sets Name field to given value.


### GetCsaw

`func (o *CVEDetailed) GetCsaw() bool`

GetCsaw returns the Csaw field if non-nil, zero value otherwise.

### GetCsawOk

`func (o *CVEDetailed) GetCsawOk() (*bool, bool)`

GetCsawOk returns a tuple with the Csaw field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCsaw

`func (o *CVEDetailed) SetCsaw(v bool)`

SetCsaw sets Csaw field to given value.


### GetPackageState

`func (o *CVEDetailed) GetPackageState() []CVEDetailedPackageState`

GetPackageState returns the PackageState field if non-nil, zero value otherwise.

### GetPackageStateOk

`func (o *CVEDetailed) GetPackageStateOk() (*[]CVEDetailedPackageState, bool)`

GetPackageStateOk returns a tuple with the PackageState field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPackageState

`func (o *CVEDetailed) SetPackageState(v []CVEDetailedPackageState)`

SetPackageState sets PackageState field to given value.

### HasPackageState

`func (o *CVEDetailed) HasPackageState() bool`

HasPackageState returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


