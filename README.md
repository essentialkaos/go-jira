<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-jira.svg"/></a></p>

<p align="center">
  <a href="https://pkg.go.dev/github.com/essentialkaos/go-jira"><img src="https://pkg.go.dev/badge/github.com/essentialkaos/go-jira"></a>
  <a href="https://goreportcard.com/report/github.com/essentialkaos/go-jira"><img src="https://goreportcard.com/badge/github.com/essentialkaos/go-jira"></a>
  <a href="https://github.com/essentialkaos/go-jira/actions"><img src="https://github.com/essentialkaos/go-jira/workflows/CI/badge.svg" alt="GitHub Actions Status" /></a>
  <a href="https://github.com/essentialkaos/go-jira/actions?query=workflow%3ACodeQL"><img src="https://github.com/essentialkaos/go-jira/workflows/CodeQL/badge.svg" /></a>
  <a href="https://codebeat.co/projects/github-com-essentialkaos-go-jira-master"><img alt="codebeat badge" src="https://codebeat.co/badges/29517531-a03f-41a5-8ef3-e77c8867d6d9" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#compatibility">Compatibility</a> • <a href="#usage-example">Usage example</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<br/>

`go-jira` is a Go package for working with [Jira REST API](https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/).

Currently, this package support only getting data from API (_i.e., you cannot create or modify data using this package_).

_**Note, that this is beta software, so it's entirely possible that there will be some significant bugs. Please report bugs so that we are aware of the issues.**_

### Installation

Make sure you have a working Go 1.14+ workspace (_[instructions](https://golang.org/doc/install)_), then:

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
| `2.x`   | Full  | Full    | Full    | No      |

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
| `master` (_Stable_) | [![CI](https://github.com/essentialkaos/go-jira/workflows/CI/badge.svg?branch=master)](https://github.com/essentialkaos/go-jira/actions) |
| `develop` (_Unstable_) | [![CI](https://github.com/essentialkaos/go-jira/workflows/CI/badge.svg?branch=develop)](https://github.com/essentialkaos/go-jira/actions) |

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
