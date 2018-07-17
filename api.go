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
	"fmt"
	"reflect"
	"strings"
	"time"
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
	Expand string       `json:"expand"`
	Fields *IssueFields `json:"fields"`
}

// IssueFields contains all available issue fields
type IssueFields struct {
	TimeSpent                     int                        `json:"timespent"`
	TimeEstimate                  int                        `json:"timeestimate"`
	TimeOriginalEstimate          int                        `json:"timeoriginalestimate"`
	AggregateTimeSpent            int                        `json:"aggregatetimespent"`
	AggregateTimeEstimate         int                        `json:"aggregatetimeestimate"`
	AggregateTimeOriginalEstimate int                        `json:"aggregatetimeoriginalestimate"`
	WorkRatio                     int                        `json:"workratio"`
	Summary                       string                     `json:"summary"`
	Description                   string                     `json:"description"`
	Environment                   string                     `json:"environment"`
	Created                       *Date                      `json:"created"`
	DueDate                       *Date                      `json:"duedate"`
	LastViewed                    *Date                      `json:"lastViewed"`
	ResolutionDate                *Date                      `json:"resolutiondate"`
	Updated                       *Date                      `json:"updated"`
	Creator                       *User                      `json:"creator"`
	Reporter                      *User                      `json:"reporter"`
	Assignee                      *User                      `json:"assignee"`
	AggregateProgress             *Progress                  `json:"aggregateprogress"`
	Progress                      *Progress                  `json:"progress"`
	IssueType                     *IssueType                 `json:"issuetype"`
	Parent                        *Issue                     `json:"parent"`
	Project                       *Project                   `json:"project"`
	Resolution                    *Resolution                `json:"resolution"`
	TimeTracking                  *TimeTracking              `json:"timetracking"`
	Watches                       *Watches                   `json:"watches"`
	Priority                      *Priority                  `json:"priority"`
	Comments                      *CommentCollection         `json:"comment"`
	Worklogs                      *WorklogCollection         `json:"worklog"`
	Votes                         *VotesInfo                 `json:"votes"`
	Status                        *Status                    `json:"status"`
	Labels                        []string                   `json:"labels"`
	Components                    []*Component               `json:"components"`
	Attachments                   []*Attachment              `json:"attachment"`
	SubTasks                      []*Issue                   `json:"subtasks"`
	Versions                      []*Version                 `json:"versions"`
	FixVersions                   []*Version                 `json:"fixVersions"`
	Custom                        map[string]json.RawMessage `json:"-"`
}

// IssueType contains info about issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	AvatarID    int    `json:"avatarId"`
	IsSubTask   bool   `json:"subtask"`
}

// Priority contains priority info
type Priority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrl"`
}

// Resolution contains resolution info
type Resolution struct {
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
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Progress contains info about issue progress
type Progress struct {
	Percent  float64 `json:"percent"`
	Progress int     `json:"progress"`
	Total    int     `json:"total"`
}

// Avatars contains avatars urls
type Avatars struct {
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

// LINKS //////////////////////////////////////////////////////////////////////////// //

// Link contains info about link
type Link struct {
	ID          string    `json:"id"`
	Type        *LinkType `json:"type"`
	InwardIssue *Issue    `json:"inwardIssue"`
}

// LinkType contains info about link type
type LinkType struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Inward  string `json:"inward"`
	Outward string `json:"outward"`
}

// RemoteLinkParams params for fetching remote link info
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

// META ///////////////////////////////////////////////////////////////////////////// //

// IssueMeta contains meta data for editing an issue
type IssueMeta struct {
	Fields map[string]*FieldMeta `json:"fields"`
}

// FieldMeta contains field meta
type FieldMeta struct {
	Required        bool              `json:"required"`
	Name            string            `json:"name"`
	Operations      []string          `json:"operations"`
	AutoCompleteURL string            `json:"autoCompleteUrl"`
	AllowedValues   []*FieldMetaValue `json:"allowedValues"`
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
	ID         string           `json:"id"`
	Name       string           `json:"name"`
	Key        string           `json:"key"`
	Category   *ProjectCategory `json:"projectCategory"`
	Avatars    *Avatars         `json:"avatarUrls"`
	IssueTypes []*IssueType     `json:"issueTypes"`
}

// ProjectCategory contains info about project category
type ProjectCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
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
	TransitionId string   `query:"transitionId"`
	Expand       []string `query:"expand"`
}

// Transition contains info about transistion
type Transition struct {
	ID     string                `json:"id"`
	Name   string                `json:"name"`
	To     *Status               `json:"to"`
	Fields map[string]*FieldMeta `json:"fields"`
}

// USERS //////////////////////////////////////////////////////////////////////////// //

// User contains user info
type User struct {
	Avatars     *Avatars `json:"avatarUrls"`
	Name        string   `json:"name"`
	Key         string   `json:"key"`
	Email       string   `json:"emailAddress"`
	DisplayName string   `json:"displayName"`
	TimeZone    string   `json:"timeZone"`
	Active      bool     `json:"active"`
}

// VERSIONS ///////////////////////////////////////////////////////////////////////// //

// Version contains version info
type Version struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsArchived  bool   `json:"archived"`
	IsReleased  bool   `json:"released"`
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

// ////////////////////////////////////////////////////////////////////////////////// //

// UnmarshalJSON is custom Date format unmarshaler
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

// UnmarshalJSON is custom IssueFields unmarshaler
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
			}
		}
	}

	return nil
}

// ////////////////////////////////////////////////////////////////////////////////// //

// ToQuery convert params to URL query
func (p EmptyParameters) ToQuery() string {
	return ""
}

// ToQuery convert params to URL query
func (p ExpandParameters) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery convert params to URL query
func (p IssueParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery convert params to URL query
func (p RemoteLinkParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery convert params to URL query
func (p CreateMetaParams) ToQuery() string {
	return paramsToQuery(p)
}

// ToQuery convert params to URL query
func (p TransitionsParams) ToQuery() string {
	return paramsToQuery(p)
}