package recall

import recdomain "github.com/kidyme/nexus/offline/internal/domain/recommendation"

func cloneCandidates(items []recdomain.Candidate, limit int) []recdomain.Candidate {
	if limit <= 0 || limit >= len(items) {
		return append([]recdomain.Candidate(nil), items...)
	}
	return append([]recdomain.Candidate(nil), items[:limit]...)
}
