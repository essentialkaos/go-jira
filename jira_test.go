package jira

// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"
	"time"

	. "pkg.re/check.v1"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type JiraSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&JiraSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *JiraSuite) TestParamsEncoding(c *C) {
	var p Parameters

	p = ExpandParameters{
		Expand: []string{"test1,test2"},
	}

	c.Assert(p.ToQuery(), Equals, `expand=test1%2Ctest2`)

	p = IssuePickerParams{
		Query:        "ABCD",
		ShowSubTasks: true,
	}

	c.Assert(p.ToQuery(), Equals, `query=ABCD&showSubTasks=true&showSubTaskParent=false`)

	p = GroupUserPickerParams{
		Query:      "ABCD",
		ShowAvatar: true,
		ProjectID:  []string{"1", "2"},
	}

	c.Assert(p.ToQuery(), Equals, `projectId=1&projectId=2&query=ABCD&showAvatar=true`)

	p = SearchParams{
		JQL:                    "ABCD",
		StartAt:                1,
		DisableQueryValidation: true,
		Expand:                 []string{"test1,test2"},
	}

	c.Assert(p.ToQuery(), Equals, `expand=test1%2Ctest2&jql=ABCD&startAt=1&validateQuery=false`)
}

func (s *JiraSuite) TestCustomUnmarshalers(c *C) {
	var err error

	d := &Date{}
	err = d.UnmarshalJSON([]byte("\"2018-05-16T23:55:39.246+0300\""))
	c.Assert(err, IsNil)
	c.Assert(d.Year(), Equals, 2018)
	c.Assert(d.Month(), Equals, time.Month(5))
	c.Assert(d.Day(), Equals, 16)
	c.Assert(d.Hour(), Equals, 23)
	c.Assert(d.Minute(), Equals, 55)
	c.Assert(d.Second(), Equals, 39)

	d = &Date{}
	err = d.UnmarshalJSON([]byte("\"2018-10-18\""))
	c.Assert(err, IsNil)
	c.Assert(d.Year(), Equals, 2018)
	c.Assert(d.Month(), Equals, time.Month(10))
	c.Assert(d.Day(), Equals, 18)

	f := &IssueFields{}
	err = f.UnmarshalJSON([]byte(`{"timespent":7200,"customfield_10700":"TEST123","resolutiondate":"2018-03-26T17:37:29.805+0300"}`))
	c.Assert(err, IsNil)
	c.Assert(f.TimeSpent, Equals, 7200)
	c.Assert(f.ResolutionDate.Day(), Equals, 26)
	c.Assert(f.Custom, HasLen, 1)
	c.Assert(f.Custom["customfield_10700"], NotNil)
}
