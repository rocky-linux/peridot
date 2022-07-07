# \DefaultApi

All URIs are relative to *https://access.redhat.com/hydra/rest/securitydata*

Method | HTTP request | Description
------------- | ------------- | -------------
[**GetCve**](DefaultApi.md#GetCve) | **Get** /cve/{CVE}.json | Get specific CVE
[**GetCves**](DefaultApi.md#GetCves) | **Get** /cve.json | Get CVEs



## GetCve

> CVEDetailed GetCve(ctx, cVE).Execute()

Get specific CVE



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    openapiclient "./openapi"
)

func main() {
    cVE := "cVE_example" // string | 

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.GetCve(context.Background(), cVE).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetCve``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetCve`: CVEDetailed
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.GetCve`: %v\n", resp)
}
```

### Path Parameters


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
**ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
**cVE** | **string** |  | 

### Other Parameters

Other parameters are passed through a pointer to a apiGetCveRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


### Return type

[**CVEDetailed**](CVEDetailed.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)


## GetCves

> []CVE GetCves(ctx).Before(before).After(after).Ids(ids).Bug(bug).Advisory(advisory).Severity(severity).Package_(package_).Product(product).Cwe(cwe).CvssScore(cvssScore).Cvss3Score(cvss3Score).Page(page).PerPage(perPage).CreatedDaysAgo(createdDaysAgo).Execute()

Get CVEs



### Example

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"
    openapiclient "./openapi"
)

func main() {
    before := time.Now() // string | CVEs before the query date. [ISO 8601 is the expected format] (optional)
    after := time.Now() // string | CVEs after the query date. [ISO 8601 is the expected format] (optional)
    ids := "ids_example" // string | CVEs for Ids separated by comma (optional)
    bug := "bug_example" // string | CVEs for Bugzilla Ids (optional)
    advisory := "advisory_example" // string | CVEs for advisory (optional)
    severity := "severity_example" // string | CVEs for severity (optional)
    package_ := "package__example" // string | CVEs which affect the package (optional)
    product := "product_example" // string | CVEs which affect the product. The parameter supports Perl compatible regular expressions. (optional)
    cwe := "cwe_example" // string | CVEs with CWE (optional)
    cvssScore := "cvssScore_example" // string | CVEs with CVSS score greater than or equal to this value (optional)
    cvss3Score := "cvss3Score_example" // string | CVEs with CVSSv3 score greater than or equal to this value (optional)
    page := float32(8.14) // float32 | CVEs for page number (optional)
    perPage := float32(8.14) // float32 | Number of CVEs to return per page (optional)
    createdDaysAgo := float32(8.14) // float32 | Index of CVEs definitions created days ago (optional)

    configuration := openapiclient.NewConfiguration()
    api_client := openapiclient.NewAPIClient(configuration)
    resp, r, err := api_client.DefaultApi.GetCves(context.Background()).Before(before).After(after).Ids(ids).Bug(bug).Advisory(advisory).Severity(severity).Package_(package_).Product(product).Cwe(cwe).CvssScore(cvssScore).Cvss3Score(cvss3Score).Page(page).PerPage(perPage).CreatedDaysAgo(createdDaysAgo).Execute()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error when calling `DefaultApi.GetCves``: %v\n", err)
        fmt.Fprintf(os.Stderr, "Full HTTP response: %v\n", r)
    }
    // response from `GetCves`: []CVE
    fmt.Fprintf(os.Stdout, "Response from `DefaultApi.GetCves`: %v\n", resp)
}
```

### Path Parameters



### Other Parameters

Other parameters are passed through a pointer to a apiGetCvesRequest struct via the builder pattern


Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **before** | **string** | CVEs before the query date. [ISO 8601 is the expected format] | 
 **after** | **string** | CVEs after the query date. [ISO 8601 is the expected format] | 
 **ids** | **string** | CVEs for Ids separated by comma | 
 **bug** | **string** | CVEs for Bugzilla Ids | 
 **advisory** | **string** | CVEs for advisory | 
 **severity** | **string** | CVEs for severity | 
 **package_** | **string** | CVEs which affect the package | 
 **product** | **string** | CVEs which affect the product. The parameter supports Perl compatible regular expressions. | 
 **cwe** | **string** | CVEs with CWE | 
 **cvssScore** | **string** | CVEs with CVSS score greater than or equal to this value | 
 **cvss3Score** | **string** | CVEs with CVSSv3 score greater than or equal to this value | 
 **page** | **float32** | CVEs for page number | 
 **perPage** | **float32** | Number of CVEs to return per page | 
 **createdDaysAgo** | **float32** | Index of CVEs definitions created days ago | 

### Return type

[**[]CVE**](CVE.md)

### Authorization

No authorization required

### HTTP request headers

- **Content-Type**: Not defined
- **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints)
[[Back to Model list]](../README.md#documentation-for-models)
[[Back to README]](../README.md)

