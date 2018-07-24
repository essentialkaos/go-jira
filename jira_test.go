package jira

// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"testing"

	. "pkg.re/check.v1"
)

// ////////////////////////////////////////////////////////////////////////////////// //

func Test(t *testing.T) { TestingT(t) }

type ConfluenceSuite struct{}

// ////////////////////////////////////////////////////////////////////////////////// //

var _ = Suite(&ConfluenceSuite{})

// ////////////////////////////////////////////////////////////////////////////////// //

func (s *ConfluenceSuite) TestParamsEncoding(c *C) {
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

	c.Assert(p.ToQuery(), Equals, `query=ABCD&showAvatar=true&projectId=1&projectId=2`)
}
