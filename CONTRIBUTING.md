# Contributing to Wasabi

First off, thank you for considering contributing to Wasabi. It's people like you that make Wasabi such a great tool.

## Where do I go from here?

If you've noticed a bug or have a feature request, make one! It's generally best if you get confirmation of your bug or approval for your feature request this way before starting to code.

## Fork & create a branch

If this is something you think you can fix, then fork Wasabi and create a branch with a descriptive name.

A good branch name would be (where issue #123 is the ticket you're working on):

```
git checkout -b 123-add-contributors-guidelines
```

## Install the project
To help you with your development environment, you can install some tools with
```
make install
```

## Get the test suite running

Make sure you're using the latest version of Go. Then, run the test suite to ensure everything is working correctly:

```
make test
```

## Get the code linter running

Install recomended version of golangci-lint 1.55.2, by following gudelines on [this page](https://golangci-lint.run/welcome/install/#local-installation).

To run golangci-lint localy, you can run command:

```
make lint
```

## Get the mock generator running

Install recomended verison of mockery v2.42.1, by following guidelines on [this page](https://vektra.github.io/mockery/latest/installation/)

To run generation of mocks, you can use command:

```
make mocks
```

## Get field aligment running

The linter has enabled rule for field alignment, to simplify fixing alignment errors it's recomended to use fieldaligment tool.

```
go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
```

## Implement your fix or feature

At this point, you're ready to make your changes! Feel free to ask for help; everyone is a beginner at first.

## Make a Pull Request

At this point, you should switch back to your master branch and make sure it's up to date with Wasabi's master branch:

```
git remote add upstream git@github.com:ksysoev/wasabi.git
git checkout master
git pull upstream master
```

Then update your feature branch from your local copy of master, and push it!

```
git checkout 123-add-contributors-guidelines
git rebase master
git push --set-upstream origin 123-add-contributors-guidelines
```

Finally, go to GitHub and make a Pull Request.

## Keeping your Pull Request updated

If a maintainer asks you to "rebase" your PR, they're saying that a lot of code has changed, and that you need to update your branch so it's easier to merge.

## Merging a PR (maintainers only)

A PR can only be merged into master by a maintainer if:

- It is passing CI.
- It has been approved by at least one maintainer. If it was a maintainer who opened the PR, only an additional maintainer can approve it.
- It has requested changes.
- It is up to date with current master.

Any maintainer is allowed to merge a PR if all of these conditions are met.
