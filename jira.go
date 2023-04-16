package jira

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2022 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/valyala/fasthttp"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// API is Jira API struct
type API struct {
	Client *fasthttp.Client // Client is client for http requests

	url  string // Jira URL
	auth string // Auth data
}

// ////////////////////////////////////////////////////////////////////////////////// //

// API errors
var (
	ErrEmptyURL     = errors.New("URL can't be empty")
	ErrNoPerms      = errors.New("User does not have permission to use Jira")
	ErrInvalidInput = errors.New("Input is invalid")
	ErrWrongLinkID  = errors.New("LinkId is not a valid number, or the remote issue link with the given id does not belong to the given issue")
	ErrNoAuth       = errors.New("Calling user is not authenticated")
	ErrNoContent    = errors.New("There is no content with the given ID, or the calling user does not have permission to view the content")
	ErrGenResponse  = errors.New("Error occurs while generating the response")
)

// ////////////////////////////////////////////////////////////////////////////////// //

// NewAPI create new API struct
func NewAPI(url string, auth Auth) (*API, error) {
	if url == "" {
		return nil, ErrEmptyURL
	}

	err := auth.Validate()

	if err != nil {
		return nil, err
	}

	return &API{
		Client: &fasthttp.Client{
			Name:                getUserAgent("", ""),
			MaxIdleConnDuration: 5 * time.Second,
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        10 * time.Second,
			MaxConnsPerHost:     150,
		},

		url:  url,
		auth: auth.Encode(),
	}, nil
}

// SetUserAgent set user-agent string based on app name and version
func (api *API) SetUserAgent(app, version string) {
	api.Client.Name = getUserAgent(app, version)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// GetConfiguration returns the information if the optional features in JIRA are
// enabled or disabled. If the time tracking is enabled, it also returns the detailed
// information about time tracking configuration.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3775
func (api *API) GetConfiguration() (*Configuration, error) {
	result := &Configuration{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/configuration",
		EmptyParameters{}, result, nil, false,
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
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetServerInfo returns general information about the current JIRA server
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4836
func (api *API) GetServerInfo(DoHealthCheck bool) (*ServerInfo, error) {
	url := "/rest/api/2/serverInfo"

	if DoHealthCheck {
		url += "?doHealthCheck=true"
	}

	result := &ServerInfo{}
	statusCode, err := api.doRequest(
		"GET", url,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetColumns returns the default system columns for issue navigator. Admin permission
// will be required.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3809
func (api *API) GetColumns() ([]*Column, error) {
	result := []*Column{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/settings/columns",
		EmptyParameters{}, &result, nil, false,
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
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetDashboards returns a list of all dashboards, optionally filtering them
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1593
func (api *API) GetDashboards(params DashboardParams) (*DashboardCollection, error) {
	result := &DashboardCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/dashboard",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetDashboard returns a single dashboard
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1621
func (api *API) GetDashboard(dashboardID string) (*Dashboard, error) {
	result := &Dashboard{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/dashboard/"+dashboardID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetFields returns a list of all fields, both System and Custom
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5959
func (api *API) GetFields() ([]*Field, error) {
	result := []*Field{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/field",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetFilter returns a filter given an id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2491
func (api *API) GetFilter(filterID string, params ExpandParameters) (*Filter, error) {
	result := &Filter{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/filter/"+filterID,
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetFilterDefaultScope returns the default share scope of the logged-in user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2616
func (api *API) GetFilterDefaultScope() (string, error) {
	result := &struct {
		Scope string `json:"scope"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/filter/defaultShareScope",
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return "", err
	}

	switch statusCode {
	case 200:
		return result.Scope, nil
	case 400:
		return "", ErrGenResponse
	case 401:
		return "", ErrNoAuth
	default:
		return "", makeUnknownError(statusCode)
	}
}

// GetFilterFavourites returns the favourite filters of the logged-in user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2597
func (api *API) GetFilterFavourites(params ExpandParameters) ([]*Filter, error) {
	result := []*Filter{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/filter/favourite",
		params, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssue returns a full representation of the issue for the given issue key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4164
func (api *API) GetIssue(issueIDOrKey string, params IssueParams) (*Issue, error) {
	result := &Issue{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey,
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueComments returns all comments for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3930
func (api *API) GetIssueComments(issueIDOrKey string, params ExpandParameters) (*CommentCollection, error) {
	result := &CommentCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/comment",
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueComment returns comment for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3987
func (api *API) GetIssueComment(issueIDOrKey, commentID string, params ExpandParameters) (*Comment, error) {
	result := &Comment{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/comment/"+commentID,
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
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
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueRemoteLinks returns sub-resource representing the remote issue links on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4385
func (api *API) GetIssueRemoteLinks(issueIDOrKey string, params RemoteLinkParams) ([]*RemoteLink, error) {
	result := []*RemoteLink{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/remotelink",
		params, &result, nil, true,
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

// GetIssueRemoteLink returns remote issue link with the given id on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4478
func (api *API) GetIssueRemoteLink(issueIDOrKey, linkID string) (*RemoteLink, error) {
	result := &RemoteLink{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/remotelink/"+linkID,
		EmptyParameters{}, result, nil, true,
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

// GetIssueTransitions returns a list of the transitions possible for this issue by the current user,
// along with fields that are required and their types
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4051
func (api *API) GetIssueTransitions(issueIDOrKey string, params TransitionsParams) ([]*Transition, error) {
	result := &struct {
		Transitions []*Transition `json:"transitions"`
	}{}

	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/transitions",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Transitions, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueVotes returns sub-resource representing the voters on the issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4143
func (api *API) GetIssueVotes(issueIDOrKey string) (*VotesInfo, error) {
	result := &VotesInfo{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/votes",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueWatchers returns the list of watchers for the issue with the given key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4232
func (api *API) GetIssueWatchers(issueIDOrKey string) (*WatchersInfo, error) {
	result := &WatchersInfo{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/watchers",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueWorklogs returns all work logs for an issue
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4232
func (api *API) GetIssueWorklogs(issueIDOrKey string) (*WorklogCollection, error) {
	result := &WorklogCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/worklog",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueWorklog returns a specific worklog
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4611
func (api *API) GetIssueWorklog(issueIDOrKey, worklogID string) (*Worklog, error) {
	result := &Worklog{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/"+issueIDOrKey+"/worklog/"+worklogID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
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
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Projects, nil
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// IssuePicker returns suggested issues which match the auto-completion query for the
// user which executes this request. This REST method will check the user's history
// and the user's browsing context and select this issues, which match the query.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4093
func (api *API) IssuePicker(params IssuePickerParams) ([]*IssuePickerResults, error) {
	result := &struct {
		Sections []*IssuePickerResults `json:"sections"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issue/picker",
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Sections, nil
	case 401:
		return nil, ErrNoAuth
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
	return api.getEntityProperties("/rest/api/2/issue/" + issueIDOrKey + "/properties")
}

// SetIssueProperty sets the value of the specified issue's property
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4889
func (api *API) SetIssueProperty(issueIDOrKey string, prop *Property) error {
	return api.setEntityProperty("/rest/api/2/issue/"+issueIDOrKey+"/properties/"+prop.Key, prop)
}

// GetIssueProperty returns the value of the property with a given key from the issue
// identified by the key or by the id. The user who retrieves the property is
// required to have permissions to read the issue.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4911
func (api *API) GetIssueProperty(issueIDOrKey, propKey string) (*Property, error) {
	return api.getEntityProperty("/rest/api/2/issue/"+issueIDOrKey+"/properties/"+propKey, propKey)
}

// DeleteIssueProperty removes the property from the issue identified by the key
// or by the id. The user removing the property is required to have permissions
// to edit the issue.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4937
func (api *API) DeleteIssueProperty(issueIDOrKey, propKey string) error {
	return api.deleteEntityProperty("/rest/api/2/issue/"+issueIDOrKey+"/properties/"+propKey, propKey)
}

// GetIssueLink returns an issue link with the specified id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3334
func (api *API) GetIssueLink(linkID string) (*Link, error) {
	result := &Link{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issueLink/"+linkID,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	case 403:
		return nil, ErrNoPerms
	case 404:
		return nil, ErrNoContent
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueLinkTypes returns a list of available issue link types, if issue
// linking is enabled. Each issue link type has an id, a name and a label
// for the outward and inward link relationship.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4959
func (api *API) GetIssueLinkTypes() ([]*LinkType, error) {
	result := &struct {
		IssueLinkTypes []*LinkType `json:"issueLinkTypes"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issueLinkType",
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.IssueLinkTypes, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueLinkType returns for a given issue link type id all information about
// this issue link type
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5004
func (api *API) GetIssueLinkType(linkTypeID string) (*LinkType, error) {
	result := &LinkType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issueLinkType/"+linkTypeID,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueTypes returns a list of all issue types visible to the user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5567
func (api *API) GetIssueTypes() ([]*IssueType, error) {
	result := []*IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issuetype",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueType returns a full representation of the issue type that has the given id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5585
func (api *API) GetIssueType(issueTypeID string) (*IssueType, error) {
	result := &IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issuetype/"+issueTypeID,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetIssueTypeAlternatives returns a list of all alternative issue types for
// the given issue type id. The list will contain these issues types, to which
// issues assigned to the given issue type can be migrated. The suitable alternatives
// are issue types which are assigned to the same workflow, the same field configuration
// and the same screen scheme.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5754
func (api *API) GetIssueTypeAlternatives(issueTypeID string) ([]*IssueType, error) {
	result := []*IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issuetype/"+issueTypeID+"/alternatives",
		EmptyParameters{}, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetAutocompleteData returns the auto complete data required for JQL searches
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1819
func (api *API) GetAutocompleteData() (*AutocompleteData, error) {
	result := &AutocompleteData{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/jql/autocompletedata",
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetAutocompleteSuggestions returns auto complete suggestions for JQL search
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1840
func (api *API) GetAutocompleteSuggestions(params SuggestionParams) ([]Suggestion, error) {
	result := &struct {
		Result []Suggestion `json:"results"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/jql/autocompletedata/suggestions",
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetMyPermissions returns all permissions in the system and whether the currently
// logged in user has them. You can optionally provide a specific context to get
// permissions for (projectKey OR projectId OR issueKey OR issueId)
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5925
func (api *API) GetMyPermissions(params PermissionsParams) (map[string]*Permission, error) {
	result := &struct {
		Permissions map[string]*Permission `json:"permissions"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/mypermissions",
		params, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Permissions, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetMyself returns currently logged user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1107
func (api *API) GetMyself() (*User, error) {
	result := &User{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/myself",
		ExpandParameters{[]string{"groups"}}, result, nil, false,
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

// GetPriorities returns a list of all issue priorities
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2804
func (api *API) GetPriorities() ([]*Priority, error) {
	result := []*Priority{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/priority",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetPriority returns an issue priority
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2822
func (api *API) GetPriority(priorityID string) (*Priority, error) {
	result := &Priority{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/priority/"+priorityID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjects returns all projects which are visible for the currently
// logged in user. If no user is logged in, it returns the list of projects
// that are visible when using anonymous access.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3484
func (api *API) GetProjects(params ExpandParameters) ([]*Project, error) {
	result := []*Project{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project",
		params, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProject returns a full representation of a project
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3509
func (api *API) GetProject(projectIDOrKey string, params ExpandParameters) (*Project, error) {
	result := &Project{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey,
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectAvatars returns all avatars which are visible for the currently logged
// in user. The avatars are grouped into system and custom.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3555
func (api *API) GetProjectAvatars(projectIDOrKey string) (*Avatars, error) {
	result := &Avatars{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/avatars",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectComponents returns a full representation of a the specified
// project's components
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3757
func (api *API) GetProjectComponents(projectIDOrKey string) ([]*Component, error) {
	result := []*Component{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/components",
		EmptyParameters{}, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectStatuses returns all issue types with valid status values for a project
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3534
func (api *API) GetProjectStatuses(projectIDOrKey string) ([]*IssueType, error) {
	result := []*IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/statuses",
		EmptyParameters{}, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrNoContent
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectVersions returns the keys of all properties for the project identified
// by the key or by the id
func (api *API) GetProjectVersions(projectIDOrKey string, params ExpandParameters) ([]*Version, error) {
	result := []*Version{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/versions",
		params, &result, nil, true,
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

// GetProjectVersion returns all versions for the specified project
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3723
func (api *API) GetProjectVersion(projectIDOrKey string, params VersionParams) (*VersionCollection, error) {
	result := &VersionCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/version",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectProperties returns the keys of all properties for the project identified
// by the key or by the id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e881
func (api *API) GetProjectProperties(projectIDOrKey string) ([]*Property, error) {
	return api.getEntityProperties("/rest/api/2/project/" + projectIDOrKey + "/properties")
}

// SetProjectProperty sets the value of the specified project's property
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e914
func (api *API) SetProjectProperty(projectIDOrKey string, prop *Property) error {
	return api.setEntityProperty("/rest/api/2/project/"+projectIDOrKey+"/properties/"+prop.Key, prop)
}

// GetProjectProperty returns the value of the property with a given key from the project
// identified by the key or by the id. The user who retrieves the property is required
// to have permissions to read the project.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e936
func (api *API) GetProjectProperty(projectIDOrKey, propKey string) (*Property, error) {
	return api.getEntityProperty("/rest/api/2/project/"+projectIDOrKey+"/properties/"+propKey, propKey)
}

// DeleteProjectProperty removes the property from the project identified by the key
// or by the id. The user removing the property is required to have permissions to
// administer the project.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e962
func (api *API) DeleteProjectProperty(projectIDOrKey, propKey string) error {
	return api.deleteEntityProperty("/rest/api/2/project/"+projectIDOrKey+"/properties/"+propKey, propKey)
}

// GetProjectRoles returns a list of roles in this project with links to full details
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4690
func (api *API) GetProjectRoles(projectIDOrKey string) (map[string]string, error) {
	result := make(map[string]string)
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/role",
		EmptyParameters{}, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectRole return details on a given project role
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e936
func (api *API) GetProjectRole(projectIDOrKey, roleID string) (*Role, error) {
	result := &Role{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/role/"+roleID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectCategories returns all project categories
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1411
func (api *API) GetProjectCategories() ([]*ProjectCategory, error) {
	result := []*ProjectCategory{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/projectCategory",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetProjectCategory returns a representation of a project category
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1459
func (api *API) GetProjectCategory(categoryID string) (*ProjectCategory, error) {
	result := &ProjectCategory{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/projectCategory/"+categoryID,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// ValidateProjectKey validates a project key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4795
func (api *API) ValidateProjectKey(projectKey string) error {
	result := &ErrorCollection{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/projectvalidate/key?key="+projectKey,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return err
	}

	switch statusCode {
	case 200:
		return result.Error()
	case 401:
		return ErrNoAuth
	default:
		return makeUnknownError(statusCode)
	}
}

// GetResolutions returns a list of all resolutions
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e842
func (api *API) GetResolutions() ([]*Resolution, error) {
	result := []*Resolution{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/resolution",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetResolution returns a resolution
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e860
func (api *API) GetResolution(resolutionID string) (*Resolution, error) {
	result := &Resolution{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/resolution/"+resolutionID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetRoles returns all the ProjectRoles available in JIRA. Currently this list
// is global.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2767
func (api *API) GetRoles() ([]*Role, error) {
	result := []*Role{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/role",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetRole returns a specific ProjectRole available in JIRA
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2786
func (api *API) GetRole(roleID string) (*Role, error) {
	result := &Role{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/role/"+roleID,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetStatuses returns a list of all statuses
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e53
func (api *API) GetStatuses() ([]*Status, error) {
	result := []*Status{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/status",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetStatus returns a full representation of the Status having the given id or name
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e74
func (api *API) GetStatus(statusIDOrName string) (*Status, error) {
	result := &Status{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/status/"+statusIDOrName,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetStatusCategories returns a list of all status categories
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5772
func (api *API) GetStatusCategories() ([]*StatusCategory, error) {
	result := []*StatusCategory{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/statuscategory",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetStatusCategory returns a full representation of the StatusCategory having
// the given id or key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5793
func (api *API) GetStatusCategory(caregoryIDOrName string) (*StatusCategory, error) {
	result := &StatusCategory{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/statuscategory/"+caregoryIDOrName,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetGroup returns representation for the requested group. Allows to get list of active
// users belonging to the specified group and its subgroups if "users" expand option is
// provided. You can page through users list by using indexes in expand param. For example
// to get users from index 10 to index 15 use "users[10:15]" expand value. This will
// return 6 users (if there are at least 16 users in this group). Indexes are 0-based
// and inclusive.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2890
func (api *API) GetGroup(params GroupParams) (*Group, error) {
	result := &Group{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/group",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
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

// GetUser returns a user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1869
func (api *API) GetUser(params UserParams) (*User, error) {
	result := &User{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetUserAvatars returns all avatars which are visible for the currently logged in user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2248
func (api *API) GetUserAvatars(username string) (*Avatars, error) {
	result := &Avatars{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user/avatars?username="+username,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetUserColumns returns the default columns for the given user. Admin permission
// will be required to get columns for a user other than the currently logged in user.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2400
func (api *API) GetUserColumns(username string) ([]*Column, error) {
	result := []*Column{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user/columns?username="+username,
		EmptyParameters{}, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	case 500:
		return nil, ErrGenResponse
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetUsersByPermissions eturns a list of active users that match the search string and
// have all specified permissions for the project or issue. This resource can be
// accessed by users with ADMINISTER_PROJECT permission for the project or global
// ADMIN or SYSADMIN rights.
//
func (api *API) GetUsersByPermissions(params UserPermissionParams) ([]*User, error) {
	result := []*User{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user/permission/search",
		params, &result, nil, true,
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

// UserPicker returns a list of users matching query with highlighting
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2027
func (api *API) UserPicker(params UserPickerParams) (*UserPickerResults, error) {
	result := &UserPickerResults{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user/picker",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GroupPicker returns groups with substrings matching a given query. This is mainly
// for use with the group picker, so the returned groups contain html to be used as
// picker suggestions. The groups are also wrapped in a single response object that
// also contains a header for use in the picker, specifically Showing X of Y matching
// groups. The number of groups returned is limited by the system property
// "jira.ajax.autocomplete.limit" The groups will be unique and sorted.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5199
func (api *API) GroupPicker(params GroupPickerParams) (*GroupPickerResults, error) {
	result := &GroupPickerResults{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/groups/picker",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GroupUserPicker returns a list of users and groups matching query with highlighting
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1792
func (api *API) GroupUserPicker(params GroupUserPickerParams) (*GroupUserPickerResults, error) {
	result := &GroupUserPickerResults{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/groupuserpicker",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// Search searches for issues using JQL
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1528
func (api *API) Search(params SearchParams) (*SearchResults, error) {
	result := &SearchResults{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/search",
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return result, ErrInvalidInput
	case 401:
		return result, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// SearchUsers returns a list of users that match the search string
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1990
func (api *API) SearchUsers(params UserSearchParams) ([]*User, error) {
	result := []*User{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/user/search",
		params, &result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetSecurityLevel returns a full representation of the security level that has
// the given id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4818
func (api *API) GetSecurityLevel(levelID string) (*SecurityLevel, error) {
	result := &SecurityLevel{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/securitylevel/"+levelID,
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetScreenFields returns available fields for screen. i.e ones that haven't
// already been added
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3189
func (api *API) GetScreenFields(screenID string) ([]*ScreenField, error) {
	result := []*ScreenField{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/screens/"+screenID+"/availableFields",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetScreenTabs returns a list of all tabs for the given screen
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3036
func (api *API) GetScreenTabs(screenID string, params ScreenParams) ([]*ScreenTab, error) {
	result := []*ScreenTab{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/screens/"+screenID+"/tabs",
		params, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetScreenTabFields returns all fields for a given tab
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3103
func (api *API) GetScreenTabFields(screenID, tabID string, params ScreenParams) ([]*ScreenField, error) {
	result := []*ScreenField{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/screens/"+screenID+"/tabs/"+tabID+"/fields",
		params, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrInvalidInput
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetVersion returns a project version
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5271
func (api *API) GetVersion(versionID string, params ExpandParameters) (*Version, error) {
	result := &Version{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/version/"+versionID,
		params, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetVersionRelatedCounts returns a bean containing the number of fixed in and affected
// issues for the given version
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5316
func (api *API) GetVersionRelatedCounts(versionID string) (int, int, error) {
	result := &struct {
		IssuesFixed    int `json:"issuesFixedCount"`
		IssuesAffected int `json:"issuesAffectedCount"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/version/"+versionID+"/relatedIssueCounts",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return 0, 0, err
	}

	switch statusCode {
	case 200:
		return result.IssuesFixed, result.IssuesAffected, nil
	case 401:
		return 0, 0, ErrNoAuth
	case 404:
		return 0, 0, ErrNoContent
	default:
		return 0, 0, makeUnknownError(statusCode)
	}
}

// GetVersionUnresolvedCount eturns the number of unresolved issues for the given version
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5337
func (api *API) GetVersionUnresolvedCount(versionID string) (int, error) {
	result := &struct {
		IssuesUnresolvedCount int `json:"issuesUnresolvedCount"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/version/"+versionID+"/unresolvedIssueCount",
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return 0, err
	}

	switch statusCode {
	case 200:
		return result.IssuesUnresolvedCount, nil
	case 401:
		return 0, ErrNoAuth
	case 404:
		return 0, ErrNoContent
	default:
		return 0, makeUnknownError(statusCode)
	}
}

// GetWorkflows returns all workflows
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1208
func (api *API) GetWorkflows() ([]*Workflow, error) {
	result := []*Workflow{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/workflow",
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWorkflow return workflow
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1208
func (api *API) GetWorkflow(workflowName string) (*Workflow, error) {
	result := &Workflow{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/workflow?workflowName="+esc(workflowName),
		EmptyParameters{}, result, nil, true,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWorkflowScheme returns the requested workflow scheme to the caller
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e292
func (api *API) GetWorkflowScheme(schemeID string, returnDraftIfExists bool) (*WorkflowScheme, error) {
	url := "/rest/api/2/workflowscheme/" + schemeID

	if returnDraftIfExists {
		url += "?returnDraftIfExists=true"
	}

	result := &WorkflowScheme{}
	statusCode, err := api.doRequest(
		"GET", url,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// GetWorkflowSchemeDefault returns the requested draft workflow scheme to the caller
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e184
func (api *API) GetWorkflowSchemeDefault(schemeID string, returnDraftIfExists bool) (string, error) {
	url := "/rest/api/2/workflowscheme/" + schemeID + "/default"

	if returnDraftIfExists {
		url += "?returnDraftIfExists=true"
	}

	result := &struct {
		Workflow string `json:"workflow"`
	}{}
	statusCode, err := api.doRequest(
		"GET", url,
		EmptyParameters{}, result, nil, false,
	)

	if err != nil {
		return "", err
	}

	switch statusCode {
	case 200:
		return result.Workflow, nil
	case 401:
		return "", ErrNoAuth
	case 404:
		return "", ErrNoContent
	default:
		return "", makeUnknownError(statusCode)
	}
}

// GetWorkflowSchemeWorkflows returns the workflow mappings or requested mapping to the caller
// for the passed scheme
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e423
func (api *API) GetWorkflowSchemeWorkflows(schemeID string, returnDraftIfExists bool) ([]*WorkflowInfo, error) {
	result := []*WorkflowInfo{}
	url := "/rest/api/2/workflowscheme/" + schemeID + "/workflow"

	if returnDraftIfExists {
		url += "?returnDraftIfExists=true"
	}

	statusCode, err := api.doRequest(
		"GET", url,
		EmptyParameters{}, &result, nil, false,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 401:
		return nil, ErrNoAuth
	case 404:
		return nil, ErrNoContent
	default:
		return nil, makeUnknownError(statusCode)
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //

// getEntityProperties returns all entity (issue/project) properties
func (api *API) getEntityProperties(url string) ([]*Property, error) {
	result := &struct {
		Keys []*Property `json:"keys"`
	}{}

	statusCode, err := api.doRequest("GET", url, EmptyParameters{}, result, nil, true)

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

// setEntityProperty create or update entity (issue/project) property
func (api *API) setEntityProperty(url string, prop *Property) error {
	statusCode, err := api.doRequest("PUT", url, EmptyParameters{}, nil, prop, false)

	if err != nil {
		return err
	}

	switch statusCode {
	case 200, 201:
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

// getEntityProperty returns entity (issue/project) property
func (api *API) getEntityProperty(url, propKey string) (*Property, error) {
	result := &struct {
		Value *Property `json:"value"`
	}{}

	statusCode, err := api.doRequest("GET", url, EmptyParameters{}, result, nil, true)

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

// deleteEntityProperty deletes entity (issue/project) property
func (api *API) deleteEntityProperty(url, propKey string) error {
	statusCode, err := api.doRequest("DELETE", url, EmptyParameters{}, nil, nil, false)

	if err != nil {
		return err
	}

	switch statusCode {
	case 204:
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
func (api *API) doRequest(method, uri string, params Parameters, result, body interface{}, decodeError bool) (int, error) {
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

	if statusCode != 200 && decodeError {
		return statusCode, decodeInternalError(resp.Body())
	}

	if result == nil {
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

	// Set authorization header
	if api.auth != "" {
		req.Header.Add("Authorization", api.auth)
	}

	return req
}

// ////////////////////////////////////////////////////////////////////////////////// //

// decodeInternalError decode internal Jira error
func decodeInternalError(data []byte) error {
	ec := &ErrorCollection{}
	err := json.Unmarshal(data, ec)

	if err != nil {
		return nil
	}

	return ec.Error()
}

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
