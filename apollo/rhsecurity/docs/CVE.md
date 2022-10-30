# CVE

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CVE** | **string** |  | 
**Severity** | **string** |  | 
**PublicDate** | **string** |  | 
**Advisories** | **[]string** |  | 
**Bugzilla** | **string** |  | 
**BugzillaDescription** | **string** |  | 
**CvssScore** | Pointer to **float32** |  | [optional] 
**CvssScoringVector** | Pointer to **string** |  | [optional] 
**CWE** | **string** |  | 
**AffectedPackages** | **[]string** |  | 
**ResourceUrl** | **string** |  | 
**Cvss3ScoringVector** | **string** |  | 
**Cvss3Score** | **string** |  | 

## Methods

### NewCVE

`func NewCVE(cVE string, severity string, publicDate string, advisories []string, bugzilla string, bugzillaDescription string, cWE string, affectedPackages []string, resourceUrl string, cvss3ScoringVector string, cvss3Score string, ) *CVE`

NewCVE instantiates a new CVE object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCVEWithDefaults

`func NewCVEWithDefaults() *CVE`

NewCVEWithDefaults instantiates a new CVE object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCVE

`func (o *CVE) GetCVE() string`

GetCVE returns the CVE field if non-nil, zero value otherwise.

### GetCVEOk

`func (o *CVE) GetCVEOk() (*string, bool)`

GetCVEOk returns a tuple with the CVE field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCVE

`func (o *CVE) SetCVE(v string)`

SetCVE sets CVE field to given value.


### GetSeverity

`func (o *CVE) GetSeverity() string`

GetSeverity returns the Severity field if non-nil, zero value otherwise.

### GetSeverityOk

`func (o *CVE) GetSeverityOk() (*string, bool)`

GetSeverityOk returns a tuple with the Severity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSeverity

`func (o *CVE) SetSeverity(v string)`

SetSeverity sets Severity field to given value.


### GetPublicDate

`func (o *CVE) GetPublicDate() string`

GetPublicDate returns the PublicDate field if non-nil, zero value otherwise.

### GetPublicDateOk

`func (o *CVE) GetPublicDateOk() (*string, bool)`

GetPublicDateOk returns a tuple with the PublicDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPublicDate

`func (o *CVE) SetPublicDate(v string)`

SetPublicDate sets PublicDate field to given value.


### GetAdvisories

`func (o *CVE) GetAdvisories() []string`

GetAdvisories returns the Advisories field if non-nil, zero value otherwise.

### GetAdvisoriesOk

`func (o *CVE) GetAdvisoriesOk() (*[]string, bool)`

GetAdvisoriesOk returns a tuple with the Advisories field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdvisories

`func (o *CVE) SetAdvisories(v []string)`

SetAdvisories sets Advisories field to given value.


### GetBugzilla

`func (o *CVE) GetBugzilla() string`

GetBugzilla returns the Bugzilla field if non-nil, zero value otherwise.

### GetBugzillaOk

`func (o *CVE) GetBugzillaOk() (*string, bool)`

GetBugzillaOk returns a tuple with the Bugzilla field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBugzilla

`func (o *CVE) SetBugzilla(v string)`

SetBugzilla sets Bugzilla field to given value.


### GetBugzillaDescription

`func (o *CVE) GetBugzillaDescription() string`

GetBugzillaDescription returns the BugzillaDescription field if non-nil, zero value otherwise.

### GetBugzillaDescriptionOk

`func (o *CVE) GetBugzillaDescriptionOk() (*string, bool)`

GetBugzillaDescriptionOk returns a tuple with the BugzillaDescription field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBugzillaDescription

`func (o *CVE) SetBugzillaDescription(v string)`

SetBugzillaDescription sets BugzillaDescription field to given value.


### GetCvssScore

`func (o *CVE) GetCvssScore() float32`

GetCvssScore returns the CvssScore field if non-nil, zero value otherwise.

### GetCvssScoreOk

`func (o *CVE) GetCvssScoreOk() (*float32, bool)`

GetCvssScoreOk returns a tuple with the CvssScore field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCvssScore

`func (o *CVE) SetCvssScore(v float32)`

SetCvssScore sets CvssScore field to given value.

### HasCvssScore

`func (o *CVE) HasCvssScore() bool`

HasCvssScore returns a boolean if a field has been set.

### GetCvssScoringVector

`func (o *CVE) GetCvssScoringVector() string`

GetCvssScoringVector returns the CvssScoringVector field if non-nil, zero value otherwise.

### GetCvssScoringVectorOk

`func (o *CVE) GetCvssScoringVectorOk() (*string, bool)`

GetCvssScoringVectorOk returns a tuple with the CvssScoringVector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCvssScoringVector

`func (o *CVE) SetCvssScoringVector(v string)`

SetCvssScoringVector sets CvssScoringVector field to given value.

### HasCvssScoringVector

`func (o *CVE) HasCvssScoringVector() bool`

HasCvssScoringVector returns a boolean if a field has been set.

### GetCWE

`func (o *CVE) GetCWE() string`

GetCWE returns the CWE field if non-nil, zero value otherwise.

### GetCWEOk

`func (o *CVE) GetCWEOk() (*string, bool)`

GetCWEOk returns a tuple with the CWE field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCWE

`func (o *CVE) SetCWE(v string)`

SetCWE sets CWE field to given value.


### GetAffectedPackages

`func (o *CVE) GetAffectedPackages() []string`

GetAffectedPackages returns the AffectedPackages field if non-nil, zero value otherwise.

### GetAffectedPackagesOk

`func (o *CVE) GetAffectedPackagesOk() (*[]string, bool)`

GetAffectedPackagesOk returns a tuple with the AffectedPackages field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAffectedPackages

`func (o *CVE) SetAffectedPackages(v []string)`

SetAffectedPackages sets AffectedPackages field to given value.


### GetResourceUrl

`func (o *CVE) GetResourceUrl() string`

GetResourceUrl returns the ResourceUrl field if non-nil, zero value otherwise.

### GetResourceUrlOk

`func (o *CVE) GetResourceUrlOk() (*string, bool)`

GetResourceUrlOk returns a tuple with the ResourceUrl field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResourceUrl

`func (o *CVE) SetResourceUrl(v string)`

SetResourceUrl sets ResourceUrl field to given value.


### GetCvss3ScoringVector

`func (o *CVE) GetCvss3ScoringVector() string`

GetCvss3ScoringVector returns the Cvss3ScoringVector field if non-nil, zero value otherwise.

### GetCvss3ScoringVectorOk

`func (o *CVE) GetCvss3ScoringVectorOk() (*string, bool)`

GetCvss3ScoringVectorOk returns a tuple with the Cvss3ScoringVector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCvss3ScoringVector

`func (o *CVE) SetCvss3ScoringVector(v string)`

SetCvss3ScoringVector sets Cvss3ScoringVector field to given value.


### GetCvss3Score

`func (o *CVE) GetCvss3Score() string`

GetCvss3Score returns the Cvss3Score field if non-nil, zero value otherwise.

### GetCvss3ScoreOk

`func (o *CVE) GetCvss3ScoreOk() (*string, bool)`

GetCvss3ScoreOk returns a tuple with the Cvss3Score field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCvss3Score

`func (o *CVE) SetCvss3Score(v string)`

SetCvss3Score sets Cvss3Score field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


