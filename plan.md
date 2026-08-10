1. **Import `crypto/subtle` in `cmd/truco-relay/main.go` and `internal/netp2p/host.go`**:
    - Update `cmd/truco-relay/main.go` to import `crypto/subtle`.
    - Update `internal/netp2p/host.go` to import `crypto/subtle`.
2. **Replace `==` and `!=` with `subtle.ConstantTimeCompare` for string comparisons that are sensitive**:
    - In `cmd/truco-relay/main.go`, replace comparisons of `req.HostAdminToken != sess.AdminToken`, `mem.Credential != req.PeerCredential`, and `mem.Credential != h.Credential` with `subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) != 1`.
    - In `internal/netp2p/host.go`, replace `joinMsg.Token != h.token` with `subtle.ConstantTimeCompare([]byte(joinMsg.Token), []byte(h.token)) != 1`.
    - Also search for any other direct comparisons of sensitive tokens in these files and replace them.
3. **Pre-commit checks**:
    - Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.
4. **Create PR**:
    - Submit the changes using the Sentinel PR naming convention.
