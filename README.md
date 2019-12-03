# Release Note

A tool for gathering release notes from all merged PRs between 2 commits on `master` branch.

# Usage

`go run release-note -user <githubUser> -token <githubToken> -repository <repository> -from <oldCommit>  -to <newCommit>  `

- `githubUser` = the user for which stats are collected
- `githubToken` = a token for acessing the github API
- `repository` = the repository in org/name format
- `oldCommit` = the commit sha of the last release
- `newCommit` = the newest commit to include, defaults to `HEAD`

# Development

TODO

# License

Release Note is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.txt).

    Copyright 2019

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.