# Flipull

Flipull is a tool for automating the generating of pull requests for build pipelines.  It is designed to be used with a tool like GitHub Actions or CircleCI to trigger pull request creations for other repositories as a part of a build.

The name is an homage to a Game Boy game from the 90s called [Flipull](https://en.wikipedia.org/wiki/Plotting_(video_game)).

## Features

Today, we only support a simple find and replace, but we plan to add more change types as needed.

## Usage

```
$ flipull replace --help
NAME:
   flipull replace - Find and replace text in a file

USAGE:
   flipull replace [command options] [arguments...]

OPTIONS:
   --github-token value   GitHub token for committing changes and opening pull requests [$GITHUB_TOKEN]
   --repo value           name of repository in the format owner/repo[@base-branch]
   --target-branch value  name of branch to create with changes (default: random branch name)
   --dry-run              output changed files to stdout instead of committing (default: false)
   --title value          title for pull request
   --description value    description for pull request (default: "This pull request was automatically generated by [flipull](https://github.com/bmorton/flipull).")
   --file value           path to file to replace text in
   --find value           text to find
   --replace value        text to replace
   --limit value          replacement limit (-1 for unlimited) (default: -1)
   --help, -h             show help (default: false)
```
