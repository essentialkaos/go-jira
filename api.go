package jira

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                     Copyright (c) 2009-2018 ESSENTIAL KAOS                         //
//        Essential Kaos Open Source License <https://essentialkaos.com/ekol>         //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// User permissions
const (
	PERMISSION_ADMINISTER                        = "ADMINISTER"
	PERMISSION_ASSIGN_ISSUE                      = "ASSIGN_ISSUE"
	PERMISSION_ASSIGNABLE_USER                   = "ASSIGNABLE_USER"
	PERMISSION_ATTACHMENT_DELETE_ALL             = "ATTACHMENT_DELETE_ALL"
	PERMISSION_ATTACHMENT_DELETE_OWN             = "ATTACHMENT_DELETE_OWN"
	PERMISSION_BROWSE                            = "BROWSE"
	PERMISSION_BULK_CHANGE                       = "BULK_CHANGE"
	PERMISSION_CLOSE_ISSUE                       = "CLOSE_ISSUE"
	PERMISSION_COMMENT_DELETE_ALL                = "COMMENT_DELETE_ALL"
	PERMISSION_COMMENT_DELETE_OWN                = "COMMENT_DELETE_OWN"
	PERMISSION_COMMENT_EDIT_ALL                  = "COMMENT_EDIT_ALL"
	PERMISSION_COMMENT_EDIT_OWN                  = "COMMENT_EDIT_OWN"
	PERMISSION_COMMENT_ISSUE                     = "COMMENT_ISSUE"
	PERMISSION_CREATE_ATTACHMENT                 = "CREATE_ATTACHMENT"
	PERMISSION_CREATE_ISSUE                      = "CREATE_ISSUE"
	PERMISSION_CREATE_SHARED_OBJECTS             = "CREATE_SHARED_OBJECTS"
	PERMISSION_DELETE_ISSUE                      = "DELETE_ISSUE"
	PERMISSION_EDIT_ISSUE                        = "EDIT_ISSUE"
	PERMISSION_LINK_ISSUE                        = "LINK_ISSUE"
	PERMISSION_MANAGE_GROUP_FILTER_SUBSCRIPTIONS = "MANAGE_GROUP_FILTER_SUBSCRIPTIONS"
	PERMISSION_MANAGE_WATCHER_LIST               = "MANAGE_WATCHER_LIST"
	PERMISSION_MODIFY_REPORTER                   = "MODIFY_REPORTER"
	PERMISSION_MOVE_ISSUE                        = "MOVE_ISSUE"
	PERMISSION_PROJECT_ADMIN                     = "PROJECT_ADMIN"
	PERMISSION_RESOLVE_ISSUE                     = "RESOLVE_ISSUE"
	PERMISSION_SCHEDULE_ISSUE                    = "SCHEDULE_ISSUE"
	PERMISSION_SET_ISSUE_SECURITY                = "SET_ISSUE_SECURITY"
	PERMISSION_SYSTEM_ADMIN                      = "SYSTEM_ADMIN"
	PERMISSION_USE                               = "USE"
	PERMISSION_USER_PICKER                       = "USER_PICKER"
	PERMISSION_VIEW_VERSION_CONTROL              = "VIEW_VERSION_CONTROL"
	PERMISSION_VIEW_VOTERS_AND_WATCHERS          = "VIEW_VOTERS_AND_WATCHERS"
	PERMISSION_VIEW_WORKFLOW_READONLY            = "VIEW_WORKFLOW_READONLY"
	PERMISSION_WORK_ISSUE                        = "WORK_ISSUE"
	PERMISSION_WORKLOG_DELETE_ALL                = "WORKLOG_DELETE_ALL"
	PERMISSION_WORKLOG_DELETE_OWN                = "WORKLOG_DELETE_OWN"
	PERMISSION_WORKLOG_EDIT_ALL                  = "WORKLOG_EDIT_ALL"
	PERMISSION_WORKLOG_EDIT_OWN                  = "WORKLOG_EDIT_OWN"
)

// Roles actors
const (
	ROLE_ACTOR_USER  = "atlassian-user-role-actor"
	ROLE_ACTOR_GROUP = "atlassian-group-role-actor"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Parameters is interface for params structs
type Parameters interface {
	ToQuery() string
}

// ExpandParameters is params with field expand info
type ExpandParameters struct {
	Expand []string `query:"expand"`
}

// EmptyParameters is empty parameters
type EmptyParameters struct {
	// nothing
}

// Date is RFC3339 encoded date
type Date struct {
	time.Time
}

// ErrorCollection is JIRA error struct
type ErrorCollection struct {
	ErrorMessages []string          `json:"errorMessages"`
	Errors        map[string]string `json:"errors"`
}

// AVATARS ////////////////////////////////////////////////////////////////////////// //

// Avatars contains info about project/user avatars
type Avatars struct {
	System []*Avatar `json:"system"`
	Custom []*Avatar `json:"custom"`
}

// Avatar contains info about project/user avatar
type Avatar struct {
	ID             string     `json:"id"`
	IsSystemAvatar bool       `json:"isSystemAvatar"`
	IsSelected     bool       `json:"isSelected"`
	AvatarURL      *AvatarURL `json:"urls"`
}

// AUTOCOMPLETE ///////////////////////////////////////////////////////////////////// //

// AutocompleteData contains autocomplete data
type AutocompleteData struct {
	VisibleFieldNames    []*JQLField    `json:"visibleFieldNames"`
	VisibleFunctionNames []*JQLFunction `json:"visibleFunctionNames"`
	ReservedWords        []string       `json:"jqlReservedWords"`
}

// JQLField contains info about JQL field
type JQLField struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	CfID        string   `json:"cfid"`
	Auto        string   `json:"auto"`
	Orderable   string   `json:"orderable"`
	Searchable  string   `json:"searchable"`
	Operators   []string `json:"operators"`
	Types       []string `json:"types"`
}

// JQLFunction contains info about JQL function
type JQLFunction struct {
	Value       string   `json:"value"`
	DisplayName string   `json:"displayName"`
	IsList      string   `json:"isList"`
	Types       []string `json:"types"`
}

// SuggestionParams is params for fetching suggestions
type SuggestionParams struct {
	FieldName      string `query:"fieldName"`
	FieldValue     string `query:"fieldValue"`
	PredicateName  string `query:"predicateName"`
	PredicateValue string `query:"predicateValue"`
}

// Suggestion contains suggestion info
type Suggestion struct {
	Value       string `json:"value"`
	DisplayName string `json:"displayName"`
}

// COLUMNS ////////////////////////////////////////////////////////////////////////// //

// Column contains info about column
type Column struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// CONFIGURATION //////////////////////////////////////////////////////////////////// //

// Configuration contains info about optional features
type Configuration struct {
	IsVotingEnabled           bool                       `json:"votingEnabled"`
	IsWatchingEnabled         bool                       `json:"watchingEnabled"`
	IsUnassignedIssuesAllowed bool                       `json:"unassignedIssuesAllowed"`
	IsSubTasksEnabled         bool                       `json:"subTasksEnabled"`
	IsIssueLinkingEnabled     bool                       `json:"issueLinkingEnabled"`
	IsTimeTrackingEnabled     bool                       `json:"timeTrackingEnabled"`
	IsAttachmentsEnabled      bool                       `json:"attachmentsEnabled"`
	TimeTrackingConfiguration *TimeTrackingConfiguration `json:"timeTrackingConfiguration"`
}

// TimeTrackingConfiguration contains detailed info about time tracking configuration
type TimeTrackingConfiguration struct {
	WorkingHoursPerDay float64 `json:"workingHoursPerDay"`
	WorkingDaysPerWeek float64 `json:"workingDaysPerWeek"`
	TimeFormat         string  `json:"timeFormat"`
	DefaultUnit        string  `json:"defaultUnit"`
}

// DASHBOARDS /////////////////////////////////////////////////////////////////////// //

// DashboardParams is params for fetching dashboards
type DashboardParams struct {
	Filter     string `query:"filter"`
	StartAt    int    `query:"startAt"`
	MaxResults int    `query:"maxResults"`
}

// DashboardCollection is dashboard collection
type DashboardCollection struct {
	StartAt    int          `json:"startAt"`
	MaxResults int          `json:"maxResults"`
	Total      int          `json:"total"`
	Data       []*Dashboard `json:"dashboards"`
}

// Dashboard contains info about dashboard
type Dashboard struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	View string `json:"view"`
}

// ISSUES /////////////////////////////////////////////////////////////////////////// //

// IssueParams is params for fetching issue info
type IssueParams struct {
	Fields []string `query:"fields,unwrap"`
	Expand []string `query:"expand"`
}

// Issue is basic issue struct
type Issue struct {
	ID     string       `json:"id"`
	Key    string       `json:"key"`
	Fields *IssueFields `json:"fields"`
}

// IssueFields contains all available issue fields
type IssueFields struct {
	TimeSpent                     int                `json:"timespent"`
	TimeEstimate                  int                `json:"timeestimate"`
	TimeOriginalEstimate          int                `json:"timeoriginalestimate"`
	AggregateTimeSpent            int                `json:"aggregatetimespent"`
	AggregateTimeEstimate         int                `json:"aggregatetimeestimate"`
	AggregateTimeOriginalEstimate int                `json:"aggregatetimeoriginalestimate"`
	WorkRatio                     int                `json:"workratio"`
	Summary                       string             `json:"summary"`
	Description                   string             `json:"description"`
	Environment                   string             `json:"environment"`
	Created                       *Date              `json:"created"`
	DueDate                       *Date              `json:"duedate"`
	LastViewed                    *Date              `json:"lastViewed"`
	ResolutionDate                *Date              `json:"resolutiondate"`
	Updated                       *Date              `json:"updated"`
	Creator                       *User              `json:"creator"`
	Reporter                      *User              `json:"reporter"`
	Assignee                      *User              `json:"assignee"`
	AggregateProgress             *Progress          `json:"aggregateprogress"`
	Progress                      *Progress          `json:"progress"`
	IssueType                     *IssueType         `json:"issuetype"`
	Parent                        *Issue             `json:"parent"`
	Project                       *Project           `json:"project"`
	Resolution                    *Resolution        `json:"resolution"`
	TimeTracking                  *TimeTracking      `json:"timetracking"`
	Watches                       *Watches           `json:"watches"`
	Priority                      *Priority          `json:"priority"`
	Comments                      *CommentCollection `json:"comment"`
	Worklogs                      *WorklogCollection `json:"worklog"`
	Votes                         *VotesInfo         `json:"votes"`
	Status                        *Status            `json:"status"`
	Security                      *SecurityLevel     `json:"security"`
	Labels                        []string           `json:"labels"`
	Components                    []*Component       `json:"components"`
	Attachments                   []*Attachment      `json:"attachment"`
	SubTasks                      []*Issue           `json:"subtasks"`
	Versions                      []*Version         `json:"versions"`
	FixVersions                   []*Version         `json:"fixVersions"`
	Issuelinks                    []*Link            `json:"issuelinks"`
	Custom                        CustomFieldsStore  `json:"-"`
}

// CustomFieldsStore is store for custom fields data
type CustomFieldsStore map[string]json.RawMessage

// IssueType contains info about issue type
type IssueType struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IconURL     string    `json:"iconUrl"`
	AvatarID    int       `json:"avatarId"`
	IsSubTask   bool      `json:"subtask"`
	Statuses    []*Status `json:"statuses"`
}

// Priority contains priority info
type Priority struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	IconURL     string `json:"iconUrl"`
	Description string `json:"description"`
	StatusColor string `json:"statusColor"`
}

// Resolution contains resolution info
type Resolution struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SecurityLevel contains info about security level
type SecurityLevel struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// TimeTracking contains info about time tracking
type TimeTracking struct {
	RemainingEstimate        string `json:"remainingEstimate"`
	TimeSpent                string `json:"timeSpent"`
	RemainingEstimateSeconds int    `json:"remainingEstimateSeconds"`
	TimeSpentSeconds         int    `json:"timeSpentSeconds"`
}

// Component contains info about component
type Component struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	AssigneeType        string `json:"assigneeType"`
	RealAssigneeType    string `json:"realAssigneeType"`
	Assignee            *User  `json:"assignee"`
	RealAssignee        *User  `json:"realAssignee"`
	IsAssigneeTypeValid bool   `json:"isAssigneeTypeValid"`
	Project             string `json:"project"`
	ProjectID           int    `json:"projectId"`
}

// Progress contains info about issue progress
type Progress struct {
	Percent  float64 `json:"percent"`
	Progress int     `json:"progress"`
	Total    int     `json:"total"`
}

// AvatarURL contains avatars urls
type AvatarURL struct {
	Size16 string `json:"16x16"`
	Size24 string `json:"24x24"`
	Size32 string `json:"32x32"`
	Size48 string `json:"48x48"`
}

// Attachment contains info about attachment
type Attachment struct {
	ID        string `json:"id"`
	Filename  string `json:"filename"`
	MIMEType  string `json:"mimeType"`
	Content   string `json:"content"`
	Thumbnail string `json:"thumbnail"`
	Created   *Date  `json:"created"`
	Author    *User  `json:"author"`
	Size      int    `json:"size"`
}

// Watches contains info about watches
type Watches struct {
	WatchCount int  `json:"watchCount"`
	IsWatching bool `json:"isWatching"`
}

// COMMENTS ///////////////////////////////////////////////////////////////////////// //

// CommentCollection is comment collection
type CommentCollection struct {
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Total      int        `json:"total"`
	Data       []*Comment `json:"comments"`
}

// Comment contains info about comment
type Comment struct {
	ID           string `json:"id"`
	Body         string `json:"body"`
	Created      *Date  `json:"created"`
	Updated      *Date  `json:"updated"`
	Author       *User  `json:"author"`
	UpdateAuthor *User  `json:"updateAuthor"`
}

// FILTERS ////////////////////////////////////////////////////////////////////////// //

// Filter contains info about filter
type Filter struct {
	ID               string                   `json:"id"`
	Name             string                   `json:"name"`
	Description      string                   `json:"description"`
	JQL              string                   `json:"jql"`
	ViewURL          string                   `json:"viewUrl"`
	SearchURL        string                   `json:"searchUrl"`
	IsFavourite      bool                     `json:"favourite"`
	Owner            *User                    `json:"owner"`
	SharedUsers      *UserCollection          `json:"sharedUsers"`
	Subscriptions    *FilterSubscriptions     `json:"subscriptions"`
	SharePermissions []*FilterSharePermission `json:"sharePermissions"`
}

// FilterSharePermission contains info about share permission
type FilterSharePermission struct {
	ID      int      `json:"id"`
	Type    string   `json:"type"`
	Project *Project `json:"project"`
	Group   *Group   `json:"group"`
}

// FilterSubscriptions contains info about filter subscriptions
type FilterSubscriptions struct {
	Size       int                   `json:"size"`
	MaxResults int                   `json:"max-results"`
	StartIndex int                   `json:"start-index"`
	EndIndex   int                   `json:"end-index"`
	Items      []*FilterSubscription `json:"items"`
}

// FilterSubscription contains info about filter subscription
type FilterSubscription struct {
	ID   int   `json:"id"`
	User *User `json:"user"`
}

// LINKS //////////////////////////////////////////////////////////////////////////// //

// Link contains info about link
type Link struct {
	ID           string    `json:"id"`
	Type         *LinkType `json:"type"`
	InwardIssue  *Issue    `json:"inwardIssue"`
	OutwardIssue *Issue    `json:"outwardIssue"`
}

// LinkType contains info about link type
type LinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// RemoteLinkParams is params for fetching remote link info
type RemoteLinkParams struct {
	GlobalID string `query:"globalId"`
}

// RemoteLink contains info about remote link
type RemoteLink struct {
	ID          int             `json:"id"`
	GlobalID    string          `json:"globalId"`
	Application *RemoteLinkApp  `json:"application"`
	Info        *RemoteLinkInfo `json:"object"`
}

// RemoteLinkInfo contains basic info about remote link
type RemoteLinkInfo struct {
	URL   string          `json:"url"`
	Title string          `json:"title"`
	Icon  *RemoteLinkIcon `json:"icon"`
}

// RemoteLinkApp contains info about link app
type RemoteLinkApp struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

// RemoteLinkIcon contains icon URL
type RemoteLinkIcon struct {
	URL string `json:"url16x16"`
}

// SCREENS ////////////////////////////////////////////////////////////////////////// //

// ScreenField contains info about screen field
type ScreenField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ScreenTab contains info about screen tab
type ScreenTab struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ScreenParams is params for fetching info about screen
type ScreenParams struct {
	ProjectKey string `query:"projectKey"`
}

// GROUPS /////////////////////////////////////////////////////////////////////////// //

// GroupParams is params for fetching groups info
type GroupParams struct {
	Name   string   `query:"groupname"`
	Expand []string `query:"expand"`
}

// Group contains info about user group
type Group struct {
	Name  string          `json:"name"`
	Users *UserCollection `json:"users"`
}

// META ///////////////////////////////////////////////////////////////////////////// //

// IssueMeta contains meta data for editing an issue
type IssueMeta struct {
	Fields map[string]*FieldMeta `json:"fields"`
}

// Field contains info about field
type Field struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	IsCustom     bool         `json:"custom"`
	IsOrderable  bool         `json:"orderable"`
	IsNavigable  bool         `json:"navigable"`
	IsSearchable bool         `json:"searchable"`
	ClauseNames  []string     `json:"clauseNames"`
	Schema       *FieldSchema `json:"schema"`
}

// FieldMeta contains field meta
type FieldMeta struct {
	Name            string            `json:"name"`
	AutoCompleteURL string            `json:"autoCompleteUrl"`
	Operations      []string          `json:"operations"`
	AllowedValues   []*FieldMetaValue `json:"allowedValues"`
	IsRequired      bool              `json:"required"`
}

// FieldSchema contains field schema
type FieldSchema struct {
	Type     string `json:"type"`
	Items    string `json:"items"`
	System   string `json:"system"`
	Custom   string `json:"custom"`
	CustomID int    `json:"customId"`
}

// FieldMetaValue contains field meta value
type FieldMetaValue struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// PERMISSIONS ////////////////////////////////////////////////////////////////////// //

// PermissionsParams is params for fetching parmissions info
type PermissionsParams struct {
	ProjectKey string `query:"projectKey"`
	ProjectID  string `query:"projectId"`
	IssueKey   string `query:"issueKey"`
	IssueID    string `query:"issueId"`
}

// Permission contains info about permission
type Permission struct {
	ID               string `json:"id"`
	Key              string `json:"key"`
	Name             string `json:"name"`
	Type             string `json:"type"`
	Description      string `json:"description"`
	IsHavePermission bool   `json:"havePermission"`
	IsDeprecatedKey  bool   `json:"deprecatedKey"`
}

// PROJECTS ///////////////////////////////////////////////////////////////////////// //

// CreateMetaParams params for fetching metadata for creating issues
type CreateMetaParams struct {
	ProjectIDs     []string `query:"projectIds"`
	ProjectKeys    []string `query:"projectKeys"`
	IssueTypeIDs   []string `query:"issuetypeIds"`
	IssueTypeNames []string `query:"issuetypeNames"`
}

// Project contains info about project
type Project struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Key          string            `json:"key"`
	URL          string            `json:"url"`
	AssigneeType string            `json:"assigneeType"`
	Lead         *User             `json:"lead"`
	Category     *ProjectCategory  `json:"projectCategory"`
	AvatarURL    *AvatarURL        `json:"avatarUrls"`
	ProjectKeys  []string          `json:"projectKeys"`
	IssueTypes   []*IssueType      `json:"issueTypes"`
	Versions     []*Version        `json:"versions"`
	Components   []*Component      `json:"components"`
	Roles        map[string]string `json:"roles"`
}

// ProjectCategory contains info about project category
type ProjectCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SEARCH /////////////////////////////////////////////////////////////////////////// //

// SearchParams is params for fetching search results
type SearchParams struct {
	JQL                    string   `query:"jql"`
	StartAt                int      `query:"startAt"`
	MaxResults             int      `query:"maxResults"`
	DisableQueryValidation bool     `query:"validateQuery,reverse"`
	Fields                 []string `query:"fields"`
	Expand                 []string `query:"expand"`
}

// SearchResults contains search result
type SearchResults struct {
	StartAt    int      `json:"startAt"`
	MaxResults int      `json:"maxResults"`
	Total      int      `json:"total"`
	Issues     []*Issue `json:"issues"`
}

// PROPERTY ///////////////////////////////////////////////////////////////////////// //

// Property contains info about property
type Property struct {
	Key   string            `json:"key"`
	Value map[string]string `json:"value"`
}

// ROLES //////////////////////////////////////////////////////////////////////////// //

// Role contains info about role
type Role struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Actors      []*Actor `json:"actors"`
}

// Actor contains info about role actor
type Actor struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
}

// STATUS  ////////////////////////////////////////////////////////////////////////// //

// Status contains info about issue status
type Status struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	IconURL     string          `json:"iconUrl"`
	Category    *StatusCategory `json:"statusCategory"`
}

// StatusCategory contains info about status category
type StatusCategory struct {
	ID        int    `json:"id"`
	Key       string `json:"key"`
	Name      string `json:"name"`
	ColorName string `json:"colorName"`
}

// TRANSITIONS ////////////////////////////////////////////////////////////////////// //

// TransitionsParams is params for fetching transitions info
type TransitionsParams struct {
	TransitionID string   `query:"transitionId"`
	Expand       []string `query:"expand"`
}

// Transition contains info about transition
type Transition struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	To     *Status               `json:"to"`
	Fields map[string]*FieldMeta `json:"fields"`
}

// USERS //////////////////////////////////////////////////////////////////////////// //

// UserParams is params for fetching user info
type UserParams struct {
	Username string   `query:"username"`
	Key      string   `query:"key"`
	Expand   []string `query:"expand"`
}

// UserPermissionParams is permissions for fetching users by permissions
type UserPermissionParams struct {
	Username    string   `query:"username"`
	Permissions []string `query:"permissions"`
	IssueKey    string   `query:"issueKey"`
	ProjectKey  string   `query:"projectKey"`
	StartAt     int      `query:"startAt"`
	MaxResults  int      `query:"maxResults"`
}

// UserSearchParams is permissions for searching users
type UserSearchParams struct {
	Username        string `query:"username"`
	StartAt         int    `query:"startAt"`
	MaxResults      int    `query:"maxResults"`
	IncludeInactive bool   `query:"includeInactive"`
	ExcludeActive   bool   `query:"includeActive,reverse"`
}

// UserCollection is users collection
type UserCollection struct {
	Size       int     `json:"size"`
	MaxResults int     `json:"max-results"`
	StartIndex int     `json:"start-index"`
	EndIndex   int     `json:"end-index"`
	Items      []*User `json:"items"`
}

// User contains user info
type User struct {
	AvatarURL   *AvatarURL  `json:"avatarUrls"`
	Name        string      `json:"name"`
	Key         string      `json:"key"`
	Email       string      `json:"emailAddress"`
	DisplayName string      `json:"displayName"`
	TimeZone    string      `json:"timeZone"`
	Locale      string      `json:"locale"`
	Groups      *UserGroups `json:"groups"`
	IsActive    bool        `json:"active"`
}

// UserGroups contains info about user groups
type UserGroups struct {
	Size  int      `json:"size"`
	Items []*Group `json:"items"`
}

// VERSIONS ///////////////////////////////////////////////////////////////////////// //

// VersionParams contains params for fetching version data
type VersionParams struct {
	StartAt    int      `query:"startAt"`
	MaxResults int      `query:"maxResults"`
	OrderBy    string   `query:"orderBy"`
	Expand     []string `query:"expand"`
}

// VersionCollection is version collection
type VersionCollection struct {
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Total      int        `json:"total"`
	IsLast     bool       `json:"isLast"`
	Data       []*Version `json:"values"`
}

// Version contains version info
type Version struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	UserReleaseDate string `json:"userReleaseDate"`
	IsArchived      bool   `json:"archived"`
	IsReleased      bool   `json:"released"`
	IsOverdue       bool   `json:"overdue"`
	ProjectID       int    `json:"projectId"`
	ReleaseDate     *Date  `json:"releaseDate"`
}

// VersionCounts contains info about issues counts
type VersionCounts struct {
	IssuesFixed    int `json:"issuesFixedCount"`
	IssuesAffected int `json:"issuesAffectedCount"`
}

// VOTES //////////////////////////////////////////////////////////////////////////// //

// VotesInfo contains info about votes
type VotesInfo struct {
	Votes    int     `json:"votes"`
	HasVoted bool    `json:"hasVoted"`
	Voters   []*User `json:"voters"`
}

// WATCHERS ///////////////////////////////////////////////////////////////////////// //

// WatchersInfo contains info about watchers
type WatchersInfo struct {
	IsWatching bool    `json:"isWatching"`
	WatchCount int     `json:"watchCount"`
	Watchers   []*User `json:"watchers"`
}

// WORK LOG ///////////////////////////////////////////////////////////////////////// //

// WorklogCollection is worklog collection
type WorklogCollection struct {
	StartAt    int        `json:"startAt"`
	MaxResults int        `json:"maxResults"`
	Total      int        `json:"total"`
	Worklogs   []*Worklog `json:"worklogs"`
}

// Worklog is worklog record
type Worklog struct {
	ID               string `json:"id"`
	Comment          string `json:"comment"`
	TimeSpent        string `json:"timeSpent"`
	Created          *Date  `json:"created"`
	Updated          *Date  `json:"updated"`
	Started          *Date  `json:"started"`
	Author           *User  `json:"author"`
	UpdateAuthor     *User  `json:"updateAuthor"`
	TimeSpentSeconds int    `json:"timeSpentSeconds"`
}

// PICKER /////////////////////////////////////////////////////////////////////////// //

// IssuePickerParams is params for fetching data from issue picker
type IssuePickerParams struct {
	Query             string `query:"query"`
	CurrentJQL        string `query:"currentJQL"`
	CurrentIssueKey   string `query:"currentIssueKey"`
	CurrentProjectID  string `query:"currentProjectId"`
	ShowSubTasks      bool   `query:"showSubTasks,respect"`
	ShowSubTaskParent bool   `query:"showSubTaskParent,respect"`
}

// IssuePickerResults contains issue picker response data
type IssuePickerResults struct {
	Label  string       `json:"label"`
	Sub    string       `json:"sub"`
	ID     string       `json:"id"`
	Msg    string       `json:"msg"`
	Issues []*IssueInfo `json:"issues"`
}

// IssueInfo contains simple info about issue
type IssueInfo struct {
	Key         string `json:"key"`
	KeyHTML     string `json:"keyHtml"`
	Img         string `json:"img"`
	Summary     string `json:"summary"`
	SummaryText string `json:"summaryText"`
}

// GroupPickerParams is params for fetching data from group picker
type GroupPickerParams struct {
	Query      string `query:"query"`
	Exclude    string `query:"exclude"`
	MaxResults int    `query:"maxResults"`
}

// GroupPickerResults contains group picker response data
type GroupPickerResults struct {
	Header string       `json:"header"`
	Total  int          `json:"total"`
	Groups []*GroupInfo `json:"groups"`
}

// GroupInfo contains simple info about group
type GroupInfo struct {
	Name string `json:"name"`
	HTML string `json:"html"`
}

// GroupUserPickerParams is params for fetching data from user/group picker
type GroupUserPickerParams struct {
	Query       string   `query:"query"`
	MaxResults  int      `query:"maxResults"`
	ShowAvatar  bool     `query:"showAvatar"`
	FieldID     string   `query:"fieldId"`
	ProjectID   []string `query:"projectId,unwrap"`
	IssueTypeID []string `query:"issueTypeId,unwrap"`
}

// GroupUserPickerResults contains user/group picker response data
type GroupUserPickerResults struct {
	Users  *UserPickerResults  `json:"users"`
	Groups *GroupPickerResults `json:"groups"`
}

// UserPickerParams is params for fetching data from user picker
type UserPickerParams struct {
	Query      string `query:"query"`
	Exclude    string `query:"exclude"`
	MaxResults int    `query:"maxResults"`
	ShowAvatar bool   `query:"showAvatar"`
}

// UserPickerResults contains user picker response data
type UserPickerResults struct {
	Header string      `json:"header"`
	Total  int         `json:"total"`
	Users  []*UserInfo `json:"users"`
}

// UserInfo contains simple info about user
type UserInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Key         string `json:"key"`
	HTML        string `json:"html"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ServerInfo contains info about JIRA instance
type ServerInfo struct {
	BuildDate      *Date          `json:"buildDate"`
	ServerTime     *Date          `json:"serverTime"`
	BaseURL        string         `json:"baseUrl"`
	Version        string         `json:"version"`
	SCMInfo        string         `json:"scmInfo"`
	ServerTitle    string         `json:"serverTitle"`
	VersionNumbers []int          `json:"versionNumbers"`
	BuildNumber    int            `json:"buildNumber"`
	HealthChecks   []*HealthCheck `json:"healthChecks"`
}

// HealthCheck contains info about health check
type HealthCheck struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPassed    bool   `json:"passed"`
}

// WORKFLOW ///////////////////////////////////////////////////////////////////////// //

// Workflow contains info about workflow
type Workflow struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	LastModifiedDate string `json:"lastModifiedDate"`
	LastModifiedUser string `json:"lastModifiedUser"`
	Steps            int    `json:"steps"`
	IsDefault        bool   `json:"default"`
}

// WorkflowInfo contains basic info about workflow
type WorkflowInfo struct {
	Workflow         string   `json:"workflow"`
	IssueTypes       []string `json:"issueTypes"`
	IsDefaultMapping bool     `json:"defaultMapping"`
}

// WorkflowScheme contains info about workflow scheme
type WorkflowScheme struct {
	ID                int                   `json:"id"`
	Name              string                `json:"name"`
	Description       string                `json:"description"`
	DefaultWorkflow   string                `json:"defaultWorkflow"`
	IsDraft           bool                  `json:"draft"`
	IssueTypeMappings map[string]string     `json:"issueTypeMappings"`
	IssueTypes        map[string]*IssueType `json:"issueTypes"`
}

// ////////////////////////////////////////////////////////////////////////////////// //

// nullBytes is a byte slice with "null" word
var nullBytes = []byte(`null`)

// ////////////////////////////////////////////////////////////////////////////////// //

// UnmarshalJSON is a custom Date format unmarshaler
func (d *Date) UnmarshalJSON(b []byte) error {
	var err error

	if bytes.Contains(b, []byte("T")) {
		d.Time, err = time.Parse("2006-01-02T15:04:05-0700", strings.Trim(string(b), "\""))
	} else {
		d.Time, err = time.Parse("2006-01-02", strings.Trim(string(b), "\""))
	}

	if err != nil {
		return fmt.Errorf("Cannot unmarshal Date value: %v", err)
	}

	return nil
}

// UnmarshalJSON is a custom IssueFields unmarshaler
func (f *IssueFields) UnmarshalJSON(b []byte) error {
	f.Custom = map[string]json.RawMessage{}

	objValue := reflect.ValueOf(f).Elem()
	knownFields := map[string]reflect.Value{}

	for i := 0; i != objValue.NumField(); i++ {
		propName := readField(objValue.Type().Field(i).Tag.Get("json"), 0, ',')
		knownFields[propName] = objValue.Field(i)
	}

	err := json.Unmarshal(b, &f.Custom)

	if err != nil {
		return err
	}

	for key, chunk := range f.Custom {
		if field, found := knownFields[key]; found {
			err = json.Unmarshal(chunk, field.Addr().Interface())

			if err != nil {
				return err
			}

			delete(f.Custom, key)
		} else {
			if !strings.HasPrefix(key, "customfield_") {
				delete(f.Custom, key)
			} else {
				if bytes.Equal(chunk, nullBytes) {
					delete(f.Custom, key)
				}
			}
		}
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Has returns true if custom field with given name exists in store
func (s CustomFieldsStore) Has(name string) bool {
	return s[name] != nil
}

// Unmarshal unmarshals custom field data
func (s CustomFieldsStore) Unmarshal(name string, v interface{}) error {
	if s[name] == nil {
		return errors.New("Custom field with name " + name + " does not exist")
	}

	return json.Unmarshal(s[name], v)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// Error returnsa  first error extracted from error collection
func (e *ErrorCollection) Error() error {
	if len(e.ErrorMessages) > 0 {
		return errors.New(e.ErrorMessages[0])
	}

	if len(e.Errors) > 0 {
		for _, v := range e.Errors {
			return errors.New(v)
		}
	}

	return nil
}

// ToQuery converts params to URL query
func (p EmptyParameters) ToQuery() string {
	return ""
}

// ToQuery converts params to URL query
func (p ExpandParameters) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p DashboardParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p GroupParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p IssueParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p RemoteLinkParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p CreateMetaParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p PermissionsParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p IssuePickerParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p GroupPickerParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p GroupUserPickerParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p ScreenParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p SearchParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p SuggestionParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p TransitionsParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p VersionParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p UserParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p UserPickerParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p UserPermissionParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery converts params to URL query
func (p UserSearchParams) ToQuery() string {
	return paramsToQuery(p)
}
