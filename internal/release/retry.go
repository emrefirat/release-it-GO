package release

import "release-it-go/internal/httputil"

// retryOptions tunes transient-failure retries for the GitHub and GitLab
// clients. The zero value is the production behavior; tests override Sleep
// so backoff does not slow the suite.
var retryOptions = httputil.Options{}
