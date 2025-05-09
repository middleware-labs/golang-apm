package tracker

import (
	"go.opentelemetry.io/otel/attribute"
	"os"
)

func addVCSAttributes(attributes []attribute.KeyValue) []attribute.KeyValue {
	// Read MW_VCS_REPOSITORY_URL and MW_VCS_COMMIT_SHA environment variables
	vcsRepositoryURL := os.Getenv("MW_VCS_REPOSITORY_URL")
	vcsCommitSHA := os.Getenv("MW_VCS_COMMIT_SHA")

	// Add their values as resource attributes
	if vcsRepositoryURL != "" {
		attributes = append(attributes, attribute.String("vcs.repository_url", vcsRepositoryURL))
	}
	if vcsCommitSHA != "" {
		attributes = append(attributes, attribute.String("vcs.commit_sha", vcsCommitSHA))
	}

	return attributes
}