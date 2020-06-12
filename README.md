<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-jira.svg"/></a></p>

<p align="center">
  <a href="https://godoc.org/pkg.re/essentialkaos/go-jira.v2"><img src="https://godoc.org/pkg.re/essentialkaos/go-jira.v2?status.svg"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/go-jira"><img src="https://goreportcard.com/badge/github.com/essentialkaos/go-jira"></a>
  <a href="https://travis-ci.com/essentialkaos/go-jira"><img src="https://travis-ci.com/essentialkaos/go-jira.svg"></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-go-jira-master"><img alt="codebeat badge" src="https://codebeat.co/badges/29517531-a03f-41a5-8ef3-e77c8867d6d9" /></a>
  <a href="https://essentialkaos.com/ekol"><img src="https://gh.kaos.st/ekol.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#compatibility">Compatibility</a> • <a href="#usage-example">Usage example</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<br/>

`go-jira` is a Go package for working with [Jira REST API](https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/).

Currently, this package support only getting data from API (_i.e., you cannot create or modify data using this package_).

_**Note, that this is beta software, so it's entirely possible that there will be some significant bugs. Please report bugs so that we are aware of the issues.**_

### Installation

Before the initial install allows git to use redirects for [pkg.re](https://github.com/essentialkaos/pkgre) service (_reason why you should do this described [here](https://github.com/essentialkaos/pkgre#git-support)_):

```
git config --global http.https://pkg.re.followRedirects true
```

Make sure you have a working Go 1.12+ workspace (_[instructions](https://golang.org/doc/install)_), then:

````
go get pkg.re/essentialkaos/go-jira.v2
````

For update to latest stable release, do:

```
go get -u pkg.re/essentialkaos/go-jira.v2
```

### Compatibility

| Version | `6.x` | `7.x`   | `8.x`   | `cloud` |
|---------|-------|---------|---------|---------|
| `1.x`   | Full  | Partial | Partial | No      |
| `2.x`   | Full  | Partial | Partial | No      |

### Usage example

```go
package main

import (
  "fmt"
  "pkg.re/essentialkaos/go-jira.v2"
)

func main() {
  api, err := jira.NewAPI("https://jira.domain.com", "john", "MySuppaPAssWOrd")
  api.SetUserAgent("MyApp", "1.2.3")

  if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
  }

  issue, err := api.GetIssue(
    "SAS-1956", jira.IssueParams{
      Expand: []string{"changelog"},
    },
  )

  if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
  }

  fmt.Println("%-v\n", issue)
}
```

### Build Status

| Branch     | Status |
|------------|--------|
| `master` (_Stable_) | [![Build Status](https://travis-ci.com/essentialkaos/go-jira.svg?branch=master)](https://travis-ci.com/essentialkaos/go-jira) |
| `develop` (_Unstable_) | [![Build Status](https://travis-ci.com/essentialkaos/go-jira.svg?branch=develop)](https://travis-ci.com/essentialkaos/go-jira) |

### License

[EKOL](https://essentialkaos.com/ekol)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
