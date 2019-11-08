# CONTRIBUTING

Welcome to [report Issues](https://github.com/tkestack/gpu-manager/issues) or [pull requests](https://github.com/tkestack/gpu-manager/pulls). It's recommended to read the following Contributing Guide first before contributing.

This document provides a set of best practices for open source contributions - bug reports, code submissions / pull requests, etc.

## Issues

We use Github Issues to track public bugs and feature requests.

### Due diligence

Before submitting a issue, please do the following:

* Perform **basic troubleshooting** steps:
    * Make sure you’re on the latest version. If you’re not on the most recent version, your problem may have been solved already! Upgrading is always the best first step.
    * Try older versions. If you’re already on the latest release, try rolling back a few minor versions (e.g. if on 1.7, try 1.5 or 1.6) and see if the problem goes away. This will help the devs narrow down when the problem first arose in the commit log.
    * Try switching up dependency versions. If the software in question has dependencies (other libraries, etc) try upgrading/downgrading those as well.
* Search the project’s bug/issue tracker to make sure it’s not a known issue.
* If you don’t find a pre-existing issue, consider checking with the mailing list and/or IRC channel in case the problem is non-bug-related.

### What to put in your bug report

Make sure your report gets the attention it deserves: bug reports with missing information may be ignored or punted back to you, delaying a fix. The below constitutes a bare minimum; more info is almost always better:

* What version of the core programming language interpreter/compiler are you using? For example, if it’s a Golang project, are you using Golang 1.13? Golang 1.12?
* What operating system are you on? Windows? (32-bit? 64-bit?) Mac OS X? (10.14? 10.10?) Linux? (Which distro? Which version of that distro? 32 or 64 bits?) Again, more detail is better.
* Which version or versions of the software are you using? Ideally, you followed the advice above and have ruled out (or verified that the problem exists in) a few different versions.
* How can the developers recreate the bug on their end? If possible, include a copy of your code, the command you used to invoke it, and the full output of your run (if applicable.) A common tactic is to pare down your code until a simple (but still bug-causing) “base case” remains. Not only can this help you identify problems which aren’t real bugs, but it means the developer can get to fixing the bug faster.

## Pull Requests

We strongly welcome your pull request to make TKEStack project better.

### Licensing of contributed material

Keep in mind as you contribute, that code, docs and other material submitted to open source projects are usually considered licensed under the same terms as the rest of the work.

Anything submitted to a project falls under the licensing terms in the repository’s top level LICENSE file. Per-file copyright/license headers are typically extraneous and undesirable. Please don’t add your own copyright headers to new files unless the project’s license actually requires them!

### Branch Management

There are three main branches here:

1. `master` branch.
	1. It is the latest (pre-)release branch. We use `master` for tags, with version number `1.1.0`, `1.2.0`, `1.3.0`...
	2. **Don't submit any PR on `master` branch.**
2. `dev` branch. 
	1. It is our stable developing branch. After full testing, `dev` will be merged to `master` branch for the next release.
	2. **You are recommended to submit bugfix or feature PR on `dev` branch.**
3. `hotfix` branch. 
	1. It is the latest tag version for hot fix. If we accept your pull request, we may just tag with version number `1.1.1`, `1.2.3`.
	2. **Only submit urgent PR on `hotfix` branch for next specific release.**

Normal bugfix or feature request should be submitted to `dev` branch. After full testing, we will merge them to `master` branch for the next release. 

If you have some urgent bugfixes on a published version, but the `master` branch have already far away with the latest tag version, you can submit a PR on hotfix. And it will be cherry picked to `dev` branch if it is possible.

```
master
 ↑
dev        <--- hotfix PR
 ↑ 
feature/bugfix PR
```  

### Make Pull Requests

The code team will monitor all pull request, we run some code check and test on it. After all tests passed, we will accecpt this PR. But it won't merge to `master` branch at once, which have some delay.

Before submitting a pull request, please make sure the followings are done:

1. Fork the repo and create your branch from `master` or `hotfix`.
2. Update code or documentation if you have changed APIs.
3. Add the copyright notice to the top of any new files you've added.
4. Check your code lints and checkstyles.
5. Test and test again your code.
6. Now, you can submit your pull request on `dev` or `hotfix` branch.

## Code Conventions

Use [Kubernetes Code Conventions](https://github.com/kubernetes/community/blob/master/contributors/guide/coding-conventions.md) for all projects in the TKEStack organization.

## Documentation isn’t optional

It’s not! Patches without documentation will be returned to sender. By “documentation” we mean:

* Docstrings must be created or updated for public API functions/methods/etc. (This step is optional for some bugfixes.)
* New features should ideally include updates to prose documentation, including useful example code snippets.
* All submissions should have a changelog entry crediting the contributor and/or any individuals instrumental in identifying the problem.

## Tests aren’t optional

Any bugfix that doesn’t include a test proving the existence of the bug being fixed, may be suspect. Ditto for new features that can’t prove they actually work.

We’ve found that test-first development really helps make features better architected and identifies potential edge cases earlier instead of later. Writing tests before the implementation is strongly encouraged.
