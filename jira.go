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
	ErrGenReponse        = errors.New("Error occurs while generating the response")
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
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        10 * time.Second,
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

// GetConfiguration returns the information if the optional features in JIRA are
// enabled or disabled. If the time tracking is enabled, it also returns the detailed
// information about time tracking configuration.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3775
func (api *API) GetConfiguration() (*Configuration, error) {
	result := &Configuration{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/configuration",
		EmptyParameters{}, result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 403:
		return nil, ErrNoPerms
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

// GetIssueLink returns an issue link with the specified id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3334
func (api *API) GetIssueLink(linkID string) (*Link, error) {
	result := &Link{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issueLink/"+linkID,
		EmptyParameters{}, &result, nil,
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
		return nil, ErrGenReponse
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
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.IssueLinkTypes, nil
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

// GetIssueTypes returns a list of all issue types visible to the user
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e5567
func (api *API) GetIssueTypes() ([]*IssueType, error) {
	result := []*IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/issuetype",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
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

// GetAutocompleteData returns the auto complete data required for JQL searches
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1819
func (api *API) GetAutocompleteData() (*AutocompleteData, error) {
	result := &AutocompleteData{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/jql/autocompletedata",
		EmptyParameters{}, &result, nil,
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
		return nil, ErrGenReponse
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
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Result, nil
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
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result.Permissions, nil
	case 400:
		return nil, ErrInvalidInput
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
		ExpandParameters{[]string{"groups"}}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
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
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
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

// GetProjects returns all projects which are visible for the currently
// logged in user. If no user is logged in, it returns the list of projects
// that are visible when using anonymous access.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3484
func (api *API) GetProjects(params ExpandParameters) ([]*Project, error) {
	result := []*Project{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project",
		params, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 500:
		return nil, ErrGenReponse
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
		params, &result, nil,
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

// GetProjectAvatars returns all avatars which are visible for the currently logged
// in user. The avatars are grouped into system and custom.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3555
func (api *API) GetProjectAvatars(projectIDOrKey string) (*ProjectAvatars, error) {
	result := &ProjectAvatars{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/avatars",
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
	case 500:
		return nil, ErrGenReponse
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

// GetProjectStatuses returns all issue types with valid status values for a project
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e3534
func (api *API) GetProjectStatuses(projectIDOrKey string) ([]*IssueType, error) {
	result := []*IssueType{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/statuses",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 400:
		return nil, ErrNoContent
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
		params, &result, nil,
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
		params, &result, nil,
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

// GetProjectProperties returns the keys of all properties for the project identified
// by the key or by the id
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e881
func (api *API) GetProjectProperties(projectIDOrKey string) ([]*Property, error) {
	result := &struct {
		Keys []*Property `json:"keys"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/properties",
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

// SetProjectProperty sets the value of the specified project's property
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e914
func (api *API) SetProjectProperty(projectIDOrKey string, prop *Property) error {
	statusCode, err := api.doRequest(
		"PUT", "/rest/api/2/project/"+projectIDOrKey+"/properties/"+prop.Key,
		EmptyParameters{}, nil, prop,
	)

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

// GetProjectProperty returns the value of the property with a given key from the project
// identified by the key or by the id. The user who retrieves the property is required
// to have permissions to read the project.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e936
func (api *API) GetProjectProperty(projectIDOrKey, propKey string) (*Property, error) {
	result := &struct {
		Value *Property `json:"value"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/properties/"+propKey,
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

// DeleteProjectProperty removes the property from the project identified by the key
// or by the id. Ths user removing the property is required to have permissions to
// administer the project.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e962
func (api *API) DeleteProjectProperty(projectIDOrKey, propKey string) error {
	statusCode, err := api.doRequest(
		"DELETE", "/rest/api/2/project/"+projectIDOrKey+"/properties/"+propKey,
		EmptyParameters{}, nil, nil,
	)

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

// GetProjectRoles returns a list of roles in this project with links to full details
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4690
func (api *API) GetProjectRoles(projectIDOrKey string) (map[string]string, error) {
	result := make(map[string]string)
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/role",
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

// GetProjectRole return details on a given project role
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e936
func (api *API) GetProjectRole(projectIDOrKey, roleID string) (*Role, error) {
	result := &Role{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/project/"+projectIDOrKey+"/role/"+roleID,
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

// GetProjectCategories returns all project categories
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e1411
func (api *API) GetProjectCategories() ([]*ProjectCategory, error) {
	result := []*ProjectCategory{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/projectCategory",
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
	case 500:
		return nil, ErrGenReponse
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

// ValidateProjectKey validates a project key
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e4795
func (api *API) ValidateProjectKey(projectKey string) error {
	result := &struct {
		Errors map[string]string `json:"errors"`
	}{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/projectvalidate/key?key="+projectKey,
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return err
	}

	switch statusCode {
	case 200:
		if len(result.Errors) == 0 {
			return nil
		} else {
			return errors.New(result.Errors["projectKey"])
		}
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
		EmptyParameters{}, &result, nil,
	)

	if err != nil {
		return nil, err
	}

	switch statusCode {
	case 200:
		return result, nil
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

// GetRoles returns all the ProjectRoles available in JIRA. Currently this list
// is global.
// https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/#d2e2767
func (api *API) GetRoles() ([]*Role, error) {
	result := []*Role{}
	statusCode, err := api.doRequest(
		"GET", "/rest/api/2/role",
		EmptyParameters{}, &result, nil,
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
		EmptyParameters{}, &result, nil,
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
