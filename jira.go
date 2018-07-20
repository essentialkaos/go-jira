package jira

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/erikdubbelboer/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// API is Jira API struct
type API struct {
	Client *fasthttp.Client // Client is client for http requests

	url       string // Jira URL
	basicAuth string // basic auth
}

// ////////////////////////////////////////////////////////////////////////////////// //

// API errors
var (
	ErrInitEmptyURL      = errors.New("URL can't be empty")
	ErrInitEmptyUser     = errors.New("User can't be empty")
	ErrInitEmptyPassword = errors.New("Password can't be empty")
	ErrNoPerms           = errors.New("User does not have permission to use confluence")
	ErrInvalidInput      = errors.New("Input is invalid")
	ErrWrongLinkID       = errors.New("LinkId is not a valid number, or the remote issue link with the given id does not belong to the given issue")
	ErrNoAuth            = errors.New("Calling user is not authenticated")
	ErrNoContent         = errors.New("There is no content with the given ID, or the calling user does not have permission to view the content")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// NewAPI create new API struct
func NewAPI(url, username, password string) (*API, error) {
	switch {
	case url == "":
		return nil, ErrInitEmptyURL
	case username == "":
		return nil, ErrInitEmptyUser
	case password == "":
		return nil, ErrInitEmptyPassword
	}

	return &API{
		Client: &fasthttp.Client{
			Name:                getUserAgent("", ""),
			MaxIdleConnDuration: 5 * time.Second,
			ReadTimeout:         3 * time.Second,
			WriteTimeout:        3 * time.Second,
			MaxConnsPerHost:     150,
		},

		url:       url,
		basicAuth: genBasicAuthHeader(username, password),
	}, nil
}

// SetUserAgent set user-agent string based on app name and version
func (api *API) SetUserAgent(app, version string) {
	api.Client.Name = getUserAgent(app, version)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetIssue returns a full representation of the issue for the given issue key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4164
func (api *API) GetIssue(issueIDOrKey string, params IssueParams) (*Issue, error) {
	result := &Issue{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey,
		params, result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetComments returns all comments for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3930
func (api *API) GetComments(issueIDOrKey string, params ExpandParameters) (*CommentCollection, error) {
	result := &CommentCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/comment",
		params, result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetComment returns comment for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3987
func (api *API) GetComment(issueIDOrKey, commentID string, params ExpandParameters) (*Comment, error) {
	result := &Comment{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/comment/"+commentID,
		params, result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueMeta returns the meta data for editing an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4364
func (api *API) GetIssueMeta(issueIDOrKey string) (*IssueMeta, error) {
	result := &IssueMeta{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/editmeta",
		EmptyParameters{}, result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetRemoteLinks returns sub-resource representing the remote issue links on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4385
func (api *API) GetRemoteLinks(issueIDOrKey string, params RemoteLinkParams) ([]*RemoteLink, error) {
	result := []*RemoteLink{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/remotelink",
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetRemoteLink returns remote issue link with the given id on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4478
func (api *API) GetRemoteLink(issueIDOrKey, linkID string) (*RemoteLink, error) {
	result := &RemoteLink{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/remotelink/"+linkID,
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrWrongLinkID
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetTransitions returns a list of the transitions possible for this issue by the current user,
// along with fields that are required and their types
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4051
func (api *API) GetTransitions(issueIDOrKey string, params TransitionsParams) ([]*Transition, error) {
	result := &struct {
		Transitions []*Transition `json:"transitions"`
	}{}

	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/transitions",
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Transitions, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetVotes returns sub-resource representing the voters on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4143
func (api *API) GetVotes(issueIDOrKey string) (*VotesInfo, error) {
	result := &VotesInfo{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/votes",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWatchers returns the list of watchers for the issue with the given key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4232
func (api *API) GetWatchers(issueIDOrKey string) (*WatchersInfo, error) {
	result := &WatchersInfo{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/watchers",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWorklogs returns all work logs for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4232
func (api *API) GetWorklogs(issueIDOrKey string) (*WorklogCollection, error) {
	result := &WorklogCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/worklog",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWorklog returns a specific worklog
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4611
func (api *API) GetWorklog(issueIDOrKey, worklogID string) (*Worklog, error) {
	result := &Worklog{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/worklog/"+worklogID,
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetCreateMeta returns the meta data for creating issues. This includes
// the available projects, issue types and fields, including field types
// and whether or not those fields are required. Projects will not be returned
// if the user does not have permission to create issues in that project.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4330
func (api *API) GetCreateMeta(params CreateMetaParams) ([]*Project, error) {
	result := &struct {
		Projects []*Project `json:"projects"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/createmeta",
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Projects, nil
	case 403:
		return nil, ErrNoPerms
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// Picker returns suggested issues which match the auto-completion query for the
// user which executes this request. This REST method will check the user's history
// and the user's browsing context and select this issues, which match the query.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4093
func (api *API) Picker(params PickerParams) ([]*PickerSection, error) {
	result := &struct {
		Sections []*PickerSection `json:"sections"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/picker",
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Sections, nil
	case 403:
		return nil, ErrNoPerms
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueProperties returns the keys of all properties for the issue identified by
// the key or by the id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4856
func (api *API) GetIssueProperties(issueIDOrKey string) ([]*Property, error) {
	result := &struct {
		Keys []*Property `json:"keys"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/properties",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Keys, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// SetIssueProperty sets the value of the specified issue's property
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4889
func (api *API) SetIssueProperty(issueIDOrKey string, prop *Property) error {
	statusCode, err := api.doRequest(
		"PUT", "/rest/api/2/issue/"+issueIDOrKey+"/properties/"+prop.Key,
		EmptyParameters{}, nil, prop,
	)

	if err != nil {
		return err
	}

	switch statusCode {
	case 200:
		return nil
	case 400:
		return ErrInvalidInput
	case 401:
		return ErrNoAuth
	case 403:
		return ErrNoPerms
	case 404:
		return ErrNoContent
	default:
		return makeUnknownError(statusCode)
	}
}

// GetIssueProperty returns the value of the property with a given key from the issue
// identified by the key or by the id. The user who retrieves the property is
// required to have permissions to read the issue.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4911
func (api *API) GetIssueProperty(issueIDOrKey, propKey string) (*Property, error) {
	result := &struct {
		Value *Property `json:"value"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/properties/"+propKey,
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Value, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// DeleteIssueProperty removes the property from the issue identified by the key
// or by the id. Ths user removing the property is required to have permissions
// to edit the issue.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4937
func (api *API) DeleteIssueProperty(issueIDOrKey, propKey string) error {
	statusCode, err := api.doRequest(
		"DELETE", "/rest/api/2/issue/"+issueIDOrKey+"/properties/"+propKey,
		EmptyParameters{}, nil, nil,
	)

	if err != nil {
		return err
	}

	switch statusCode {
	case 200:
		return nil
	case 400:
		return ErrInvalidInput
	case 401:
		return ErrNoAuth
	case 403:
		return ErrNoPerms
	case 404:
		return ErrNoContent
	default:
		return makeUnknownError(statusCode)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// codebeat:disable[ARITY]

// doRequest create and execute request
func (api *API) doRequest(method, uri string, params Parameters, result, body interface{}) (int, error) {
	req := api.acquireRequest(method, uri, params)
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	if body != nil {
		bodyData, err := json.Marshal(body)

		if err != nil {
			return -1, err
		}

		req.SetBody(bodyData)
	}

	err := api.Client.Do(req, resp)

	if err != nil {
		return -1, err
	}

	statusCode := resp.StatusCode()

	if statusCode != 200 || result == nil {
		return statusCode, nil
	}

	err = json.Unmarshal(resp.Body(), result)

	return statusCode, err
}

// codebeat:enable[ARITY]

// acquireRequest acquire new request with given params
func (api *API) acquireRequest(method, uri string, params Parameters) *fasthttp.Request {
	req := fasthttp.AcquireRequest()
	query := params.ToQuery()

	req.SetRequestURI(api.url + uri)

	// Set query if params can be encoded as query
	if query != "" {
		req.URI().SetQueryString(query)
	}

	if method != "GET" {
		req.Header.SetContentType("application/json")
		req.Header.SetMethod(method)
	}

	// Set auth header
	req.Header.Add("Authorization", "Basic "+api.basicAuth)

	return req
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getUserAgent generate user-agent string for client
func getUserAgent(app, version string) string {
	if app != "" && version != "" {
		return fmt.Sprintf(
			"%s/%s %s/%s (go; %s; %s-%s)",
			app, version, NAME, VERSION, runtime.Version(),
			runtime.GOARCH, runtime.GOOS,
		)
	}

	return fmt.Sprintf(
		"%s/%s (go; %s; %s-%s)",
		NAME, VERSION, runtime.Version(),
		runtime.GOARCH, runtime.GOOS,
	)
}

// genBasicAuthHeader generate basic auth header
func genBasicAuthHeader(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// makeUnknownError create error struct for unknown error
func makeUnknownError(statusCode int) error {
	return fmt.Errorf("Unknown error occurred (status code %d)", statusCode)
}
