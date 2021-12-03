# GitHubAPI

Implemented GitHub API HTTP client that:

Reads a text file given as command line argument to the program and parses different Github usernames - each username on separate line in the file.
Fetches GitHub users data in JSON format using public GitHub API: https://api.github.com/users/${username}.
Fetches GitHub user repositories data in JSON format from: https://api.github.com/users/${username}/repos.
Fetches information about programming languages in each repo from: https://api.github.com/repos/${username}/${repo-name}/languages.

A statistics report is printed containing:
 - Username
 - Number of repositories
 - Distribution of programming languages
 - Number of followers
 - Number of forks for all repositories


This is a homework project part of the Chaos GoLang Camp.
