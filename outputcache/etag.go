package outputcache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateETag generates an ETag for the given response body.
func generateETag(body []byte) string {
	hash := sha256.Sum256(body)
	return fmt.Sprintf(`"%s"`, hex.EncodeToString(hash[:16]))
}

// shouldRevalidate checks if the request includes If-None-Match and it matches the ETag.
func shouldRevalidate(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" || etag == "" {
		return false
	}
	return ifNoneMatch == etag || ifNoneMatch == "*"
}
