<p align="center"><a href="#readme"><img src="https://gh.kaos.st/go-jira.svg"/></a></p>

<p align="center">
  <a href="https://kaos.sh/g/go-jira.v2"><img src="https://gh.kaos.st/godoc.svg" alt="PkgGoDev" /></a>
  <a href="https://kaos.sh/r/go-jira"><img src="https://kaos.sh/r/go-jira.svg" alt="GoReportCard" /></a>
  <a href="https://kaos.sh/w/go-jira/ci"><img src="https://kaos.sh/w/go-jira/ci.svg" alt="GitHub Actions CI Status" /></a>
  <a href="https://kaos.sh/w/go-jira/codeql"><img src="https://kaos.sh/w/go-jira/codeql.svg" alt="GitHub Actions CodeQL Status" /></a>
  <a href="https://kaos.sh/b/go-jira"><img src="https://kaos.sh/b/29517531-a03f-41a5-8ef3-e77c8867d6d9.svg" alt="Codebeat badge" /></a>
  <a href="#license"><img src="https://gh.kaos.st/apache2.svg"></a>
</p>

<p align="center"><a href="#installation">Installation</a> • <a href="#compatibility">Compatibility</a> • <a href="#usage-example">Usage example</a> • <a href="#build-status">Build Status</a> • <a href="#license">License</a></p>

<br/>

`go-jira` is a Go package for working with [Jira REST API](https://docs.atlassian.com/software/jira/docs/api/REST/6.4.13/).

Currently, this package support only getting data from API (_i.e., you cannot create or modify data using this package_).

_**Note, that this is beta software, so it's entirely possible that there will be some significant bugs. Please report bugs so that we are aware of the issues.**_

### Installation

Make sure you have a working Go 1.15+ workspace (_[instructions](https://golang.org/doc/install)_), then:

````
go get -d pkg.re/essentialkaos/go-jira.v2
````

For update to latest stable release, do:

```
go get -d -u pkg.re/essentialkaos/go-jira.v2
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
| `master` (_Stable_) | [![CI](https://kaos.sh/w/go-jira/ci.svg?branch=master)](https://kaos.sh/w/go-jira/ci?query=branch:master) |
| `develop` (_Unstable_) | [![CI](https://kaos.sh/w/go-jira/ci.svg?branch=develop)](https://kaos.sh/w/go-jira/ci?query=branch:develop) |

### License

[Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0)

<p align="center"><a href="https://essentialkaos.com"><img src="https://gh.kaos.st/ekgh.svg"/></a></p>
