package tracker

import (
	"os"

	gogit "github.com/go-git/go-git/v5"
	"go.opentelemetry.io/otel/attribute"
)

func addVCSAttributes(attributes []attribute.KeyValue) []attribute.KeyValue {
	// Read MW_VCS_REPOSITORY_URL and MW_VCS_COMMIT_SHA environment variables
	vcsRepositoryURL := os.Getenv("MW_VCS_REPOSITORY_URL")
	vcsCommitSHA := os.Getenv("MW_VCS_COMMIT_SHA")

	// If missing, try to get from .git using go-git (pure Go, no git CLI needed)
	if vcsRepositoryURL == "" || vcsCommitSHA == "" {
		gitDir := findGitDir()
		if gitDir != "" {
			repo, err := gogit.PlainOpen(gitDir)
			if err == nil {
				if vcsCommitSHA == "" {
					head, err := repo.Head()
					if err == nil {
						vcsCommitSHA = head.Hash().String()
					}
				}
				if vcsRepositoryURL == "" {
					remotes, err := repo.Remotes()
					if err == nil && len(remotes) > 0 {
						urls := remotes[0].Config().URLs
						if len(urls) > 0 {
							vcsRepositoryURL = urls[0]
							// Remove .git suffix if present
							if len(vcsRepositoryURL) > 4 && vcsRepositoryURL[len(vcsRepositoryURL)-4:] == ".git" {
								vcsRepositoryURL = vcsRepositoryURL[:len(vcsRepositoryURL)-4]
							}
						}
					}
				}
			}
		}
	}

	// Add their values as resource attributes
	if vcsRepositoryURL != "" {
		attributes = append(attributes, attribute.String("vcs.repository_url", vcsRepositoryURL))
	}
	if vcsCommitSHA != "" {
		attributes = append(attributes, attribute.String("vcs.commit_sha", vcsCommitSHA))
	}

	return attributes
}

// Helper to find .git directory in current or parent directories
func findGitDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		gitPath := dir + "/.git"
		if stat, err := os.Stat(gitPath); err == nil && stat.IsDir() {
			return dir
		}
		parent := parentDir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// Helper to get parent directory
func parentDir(path string) string {
	if len(path) == 0 {
		return ""
	}
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			if i == 0 {
				return "/"
			}
			return path[:i]
		}
	}
	return path
}
