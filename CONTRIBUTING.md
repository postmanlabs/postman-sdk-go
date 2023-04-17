## Naming conventions
<br>

### PR Title
PR Title should start with the JIRA ticket Id in square brackets, example: [AOB-100] with a human readable title to accompany it.

<br>

### Branch Names
Branch names should start with an action followed by a concise name of the action to be performed.

E.g.
- feature/add-hostname-support
- bug/incorrect-user-handling
- hotfix/no-userid-support

<br>

### Commit message
Commit messages should be descriptive and highlight the action performed. They should also start with the type of action performed, such as feat and fix. This leads to better
documentation when creating Github Releases.

Examples:
- feat: Use distribution bot app to upload tags
- fix: Stop spans from exporting when suppressed

They should NOT be terse and context-less.

Such as:
- fixed linting issue.
- format fix.

<br>

## Release Process
- Merge all the changes in `develop`.
- Create a new branch from `develop`, update the `version.go` file with the updated version. Raise a PR, get it reviwed and merged to `develop`.
- At this point, we have all the changes in `develop` which we want to tag and create a release for. Create a PR from `develop` to `master`.
- Fast-forward merge the PR into `master` [**Do not merge via the UI**]. This will trigger the `Release Go SDK` action, which will create the tag and Github release for the same.
```shell
# On master branch
git merge --ff develop
```
- Check the progress in Actions tab, and find the newly created release under Releases tag.