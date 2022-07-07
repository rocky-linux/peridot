# CVEDetailedAffectedRelease

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ProductName** | **string** |  | 
**ReleaseDate** | **string** |  | 
**Advisory** | **string** |  | 
**Cpe** | **string** |  | 
**Package** | Pointer to **string** |  | [optional] 

## Methods

### NewCVEDetailedAffectedRelease

`func NewCVEDetailedAffectedRelease(productName string, releaseDate string, advisory string, cpe string, ) *CVEDetailedAffectedRelease`

NewCVEDetailedAffectedRelease instantiates a new CVEDetailedAffectedRelease object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCVEDetailedAffectedReleaseWithDefaults

`func NewCVEDetailedAffectedReleaseWithDefaults() *CVEDetailedAffectedRelease`

NewCVEDetailedAffectedReleaseWithDefaults instantiates a new CVEDetailedAffectedRelease object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProductName

`func (o *CVEDetailedAffectedRelease) GetProductName() string`

GetProductName returns the ProductName field if non-nil, zero value otherwise.

### GetProductNameOk

`func (o *CVEDetailedAffectedRelease) GetProductNameOk() (*string, bool)`

GetProductNameOk returns a tuple with the ProductName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProductName

`func (o *CVEDetailedAffectedRelease) SetProductName(v string)`

SetProductName sets ProductName field to given value.


### GetReleaseDate

`func (o *CVEDetailedAffectedRelease) GetReleaseDate() string`

GetReleaseDate returns the ReleaseDate field if non-nil, zero value otherwise.

### GetReleaseDateOk

`func (o *CVEDetailedAffectedRelease) GetReleaseDateOk() (*string, bool)`

GetReleaseDateOk returns a tuple with the ReleaseDate field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReleaseDate

`func (o *CVEDetailedAffectedRelease) SetReleaseDate(v string)`

SetReleaseDate sets ReleaseDate field to given value.


### GetAdvisory

`func (o *CVEDetailedAffectedRelease) GetAdvisory() string`

GetAdvisory returns the Advisory field if non-nil, zero value otherwise.

### GetAdvisoryOk

`func (o *CVEDetailedAffectedRelease) GetAdvisoryOk() (*string, bool)`

GetAdvisoryOk returns a tuple with the Advisory field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAdvisory

`func (o *CVEDetailedAffectedRelease) SetAdvisory(v string)`

SetAdvisory sets Advisory field to given value.


### GetCpe

`func (o *CVEDetailedAffectedRelease) GetCpe() string`

GetCpe returns the Cpe field if non-nil, zero value otherwise.

### GetCpeOk

`func (o *CVEDetailedAffectedRelease) GetCpeOk() (*string, bool)`

GetCpeOk returns a tuple with the Cpe field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCpe

`func (o *CVEDetailedAffectedRelease) SetCpe(v string)`

SetCpe sets Cpe field to given value.


### GetPackage

`func (o *CVEDetailedAffectedRelease) GetPackage() string`

GetPackage returns the Package field if non-nil, zero value otherwise.

### GetPackageOk

`func (o *CVEDetailedAffectedRelease) GetPackageOk() (*string, bool)`

GetPackageOk returns a tuple with the Package field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPackage

`func (o *CVEDetailedAffectedRelease) SetPackage(v string)`

SetPackage sets Package field to given value.

### HasPackage

`func (o *CVEDetailedAffectedRelease) HasPackage() bool`

HasPackage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


