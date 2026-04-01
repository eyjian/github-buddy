## 1. Candidate IP quality improvements

- [x] 1.1 Implement candidate IP validation to reject invalid, private, loopback, link-local, multicast, and other non-public addresses
- [x] 1.2 Change data source fetching from first-success failover to multi-source aggregation with deduplication
- [x] 1.3 Review and refresh the maintained domain list to ensure browser-critical GitHub domains remain covered

## 2. Browser-oriented verification

- [x] 2.1 Extend HTTPS verification so `github.com` uses stronger homepage success heuristics beyond status code only
- [x] 2.2 Preserve lightweight HTTPS verification for non-homepage GitHub domains
- [x] 2.3 Adjust best-IP selection so browser-validated `github.com` candidates are preferred, while keeping best-effort fallback behavior

## 3. User-visible diagnostics

- [x] 3.1 Update `status` output to distinguish port reachability from HTTPS/browser-level reachability
- [x] 3.2 Update `init` and `update` summaries to report `github.com` browser-oriented verification results
- [x] 3.3 Keep browser cache and DNS flush guidance aligned with the stronger diagnostics

## 4. Testing and validation

- [x] 4.1 Add unit tests for candidate IP filtering and multi-source aggregation
- [x] 4.2 Add unit tests for browser-oriented `github.com` response validation heuristics
- [x] 4.3 Add regression tests for status/selection behavior when TCP succeeds but browser-oriented validation fails
- [x] 4.4 Validate behavior with targeted `curl` or `wget` checks and run the relevant Go tests
