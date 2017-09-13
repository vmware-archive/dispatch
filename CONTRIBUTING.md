# Contributing to Serverless

This document lives in [CONTRIBUTING.md](CONTRIBUTING.md), feel free to improve and expand with a PR.

## Table Of Contents

[What should I know before I get started?](#what-should-i-know-before-i-get-started)
- [Setting up the Development Environment](#setting-up-the-development-environment)

[How Do I Get Started?](#how-do-i-get-started)
- [How to build Serverless](#how-to-build-serverless)
- [Your First Code Contribution](#your-first-code-contribution)

[Best Practices](#best-practices)
- [Git Commit Messages](#git-commit-messages)
- [Go Styleguide](#go-styleguide)
- [Swagger Styleguide](#swagger-styleguide)
- [Markdown Styleguide](#markdown-styleguide)

[Additional Notes](#additional-notes)

## What should I know before I get started?

### Setting up the Development Environment

> Before attempting to build this project, you should first read up on setting up a Go workspace on your machine: [https://golang.org/doc/code.html](https://golang.org/doc/code.html)

First, fork the repository on [gitlab](https://gitlab.eng.vmware.com/serverless/serverless) to your personal account.

Next, create a project directory, clone the repository, and setup hooks:

```shell
$ mkdir -p $HOME/Dev/serverless/src/gitlab.eng.vmware.com/serverless
$ git clone git@gitlab.eng.vmware.com:serverless/serverless.git
$ git remote add $USER git@gitlab.eng.vmware.com:$USER/serverless.git
$ git remote -v
bjung	git@gitlab.eng.vmware.com:bjung/serverless.git (fetch)
bjung	git@gitlab.eng.vmware.com:bjung/serverless.git (push)
origin	git@gitlab.eng.vmware.com:serverless/serverless.git (fetch)
origin	git@gitlab.eng.vmware.com:serverless/serverless.git (push)
$ git fetch --all
$ git config core.hooksPath .githooks
```

Now install the prerequisites for the development environment:

```shell
# Assuming you're on Mac
$ brew install go@1.9
$ brew install dep
$ brew install go-swagger
$ ./scripts/toolchain.sh
```

> Dep is used for dependency management for go.  For more info and usage go to
[https://github.com/golang/dep](https://github.com/golang/dep)

Make sure the following path is in your `PATH`.  This should be in your `.bash_profile`:

```shell
$ export PATH=$PATH:/usr/local/opt/go/libexec/bin
```

Note that `GOPATH` can be any directory, the example below uses `$HOME/Dev/serverless`.  It's not a bad idea to add
this to your `.bash_profile`.

```shell
$ export GOPATH=$HOME/Dev/serverless
```

Make sure you're not using any symlink to access this directory, scripts and checks will not work correctly.

## How Do I Get Started?

### How to build Serverless

To build Serveless you'll need a recent version of `make` and a working `go 1.9` install.

The first thing to do is pull or update dependencies:

```shell
$ dep ensure
```

To build all the Serverless artifacts, run:

```shell
# Assuming you're on Mac (otherwise `make linux`)
make darwin
```

You can run the linter check with:

```shell
make check
```

and Perform unit tests with:

```shell
make test
```

### Your First Code Contribution

This is a rough outline of what a contributor's workflow looks like:
- Create a topic branch from where you want to base your work.
- Make commits of logical units.
- Make sure your commit messages are in the proper format (see below).
- Push your changes to a topic branch in your fork of the repository.
- Submit a merge request to `serverless/serverless`.
- Your MR must receive at least one approval before merging.

Example:

```shell
$ git checkout -b my-new-feature origin/master
$ git commit -a
$ git push $USER my-new-feature
...
remote:
remote: To create a merge request for contributing, visit:
remote:   https://gitlab.eng.vmware.com/bjung/serverless/merge_requests/new?merge_request%5Bsource_branch%5D=my-new-feature
remote:
To gitlab.eng.vmware.com:bjung/serverless.git
 * [new branch]      my-new-feature -> my-new-feature
```

Note the merge request URL.  Follow that link to create a new merge request against `origin/master`.

#### Setup the CI runner on gitlab

Whenever a merge request is created (or updated) gitlab will kick off CI tests.  If the CI runners haven't been
configured for your repository (fork), the CI tests will stall in "Pending" state waiting for a runner.  To enable
a runner for your repository open [https://gitlab.eng.vmware.com/$USER/serverless/settings/ci_cd](https://gitlab.eng.vmware.com/$USER/serverless/settings/ci_cd) in your browser.  You should see a list of
"Available specific runners".  Click on the "Enable for this project" button next to one or more runners.

#### Stay in sync with upstream

When your branch gets out of sync with the `origin/master` branch, use the following to update:

```shell
git checkout my-new-feature
git fetch -all
git rebase origin/master
git push --force-with-lease $USER my-new-feature
```

If you have an open merge request, it will automatically get picked up.  There should be no need to update the merge
request in gitlab.

#### Updating pull requests

If your PR fails to pass CI or needs changes based on code review, you'll most likely want to squash these changes into existing commits.

If your pull request contains a single commit or your changes are related to the most recent commit, you can simply amend the commit.

```shell
git add .
git commit --amend
git push --force-with-lease $USER my-new-feature
```

If you need to squash changes into an earlier commit, you can use:

```shell
git add .
git commit --fixup <commit>
git rebase -i --autosquash origin/master
git push --force-with-lease $USER my-new-feature
```

## Best Practices

### Git Commit Messages
> Condensed from [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/) by Chris Beams

#### :sparkles: The 10 golden rules :sparkles:
1. Separate subject from body with a blank line
2. Limit the subject line to 50 characters
3. Capitalize the subject line
4. Do not end the subject line with a period
5. Use the present tense ("Add feature" not "Added feature")
6. Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
7. Wrap the body at 72 characters
8. Use the body to explain what and why vs. how
9. When only changing documentation, include `[ci skip]` in the commit description
10. Reference issues and pull requests liberally

### Definition of Done

In order for a feature to be considered done it meets the following criteria.
- The functionality described is implemented
- Tests that verify the functionality work and provide 80% or higher code coverage
- Documentation is updated as appropriate (e.g. changes and additions to APIs are documented clearly along with
  parameters, allowable values, semantics, errors and their meanings
- All anticipated failure modes are handled gracefully
- Error messages are clear and informative and written with the intent of guiding a customer towards the likely cause
  and corrective action

### Go Styleguide
The coding style suggested by the Golang community is used in Cello. See the [style doc](https://github.com/golang/go/wiki/CodeReviewComments) for details.

Try to limit column width to 120 characters.

#### Logging

- **FATAL** - Not all loggers support this level. Ops needs to take immediate action when this happens, this should
  translate into P0 drop everything right now and look at this.
- **ERROR** - This is very similar to fatal except that it doesn't force a restart. It should still not be taken lightly
  because again it should make people drop whatever they are doing and take a look at what's going on.
- **WARN** - Something is off, in dev and stage this should be logged in production we can ignore this. We can safely
  recover from this and continue our on merry way.
- **INFO** - Hey something interesting happened, here's what I did. This should not have lots of context like data dumps
  but should be concise descriptive messages. Info is often turned off in staging environments, but should be turned on
  in the development.
- **DEBUG** - companion to INFO but should be used to dump the data that you require when debugging, might be turned on
  dev envs, would typically be turned on test runs
- **TRACE** - enter/leave methods, everything is fair game here

### Swagger styleguide

Swaggers canonical representation is JSON. We'll use the JSON naming conventions for representing data in a swagger
specification.

Naming a swagger file:
- When your file is an API definition used for codegen: `[api-name]-[version]-swagger.yml`
- When your file is a supporting file: `[descriptive-name].yml`

#### Property Name Format

Choose meaningful property names.

Property names must conform to the following guidelines:
- Property names should be meaningful names with defined semantics.
- Property names must be camel-cased, ascii strings.
- The first character must be a letter, an underscore (`_`) or a dollar sign (`$`).
- Subsequent characters can be a letter, a digit, an underscore, or a dollar sign.
- Reserved JavaScript keywords should be avoided.

```json
{
  "thisPropertyIsAnIdentifier": "identifier value"
}
```

#### Singular vs Plural Property Names

Array types should have plural property names. All other property names should be singular. Arrays usually contain
multiple items, and a plural property name reflects this. An example of this can be seen in the reserved names below.
The items property name is plural because it represents an array of item objects. Most of the other fields are singular.

There may be exceptions to this, especially when referring to numeric property values. For example, in the reserved
names, totalItems makes more sense than totalItem. However, technically, this is not violating the style guide, since
totalItems can be viewed as totalOfItems, where total is singular (as per the style guide), and OfItems serves to
qualify the total. The field name could also be changed to itemCount to look singular.

```json
{
  "author": "lisa",
  "siblings": [ "bart", "maggie"],
  "totalItems": 10,
  "itemCount": 10,
}
```

### Markdown Styleguide

Markdown documents should be created using the conventions found in the [Carrot Creative's Markdown
Styleguide](https://github.com/carrot/markdown-styleguide), [`tidy-markdown`](https://github.com/slang800/tidy-markdown)
is an utility to fix ugly markdown, there are several editor plugins that leverage this plugin on save.

## Additional Notes
> TODO