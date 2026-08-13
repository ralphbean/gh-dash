package data

import (
	"testing"

	gh "github.com/cli/go-gh/v2/pkg/api"
	"github.com/stretchr/testify/require"
)

func TestClearEnrichmentCache(t *testing.T) {
	// Save original state
	originalCachedClient := cachedClient
	defer func() {
		cachedClient = originalCachedClient
	}()

	t.Run("clears nil cache without panic", func(t *testing.T) {
		cachedClient = nil
		require.True(t, IsEnrichmentCacheCleared(), "cache should be cleared initially")

		ClearEnrichmentCache()
		require.True(t, IsEnrichmentCacheCleared(), "cache should remain cleared")
	})

	t.Run("clears non-nil cache", func(t *testing.T) {
		// Simulate having a cached client (we use an empty struct pointer
		// since we can't create a real GraphQL client without credentials)
		cachedClient = &gh.GraphQLClient{}
		require.False(
			t,
			IsEnrichmentCacheCleared(),
			"cache should not be cleared when client is set",
		)

		ClearEnrichmentCache()
		require.True(
			t,
			IsEnrichmentCacheCleared(),
			"cache should be cleared after ClearEnrichmentCache",
		)
	})
}

func TestIsEnrichmentCacheCleared(t *testing.T) {
	// Save original state
	originalCachedClient := cachedClient
	defer func() {
		cachedClient = originalCachedClient
	}()

	t.Run("returns true when cache is nil", func(t *testing.T) {
		cachedClient = nil
		require.True(t, IsEnrichmentCacheCleared())
	})

	t.Run("returns false when cache is set", func(t *testing.T) {
		cachedClient = &gh.GraphQLClient{}
		require.False(t, IsEnrichmentCacheCleared())
	})
}

func TestUnresolvedThreadsCount(t *testing.T) {
	newNodes := func(resolved ...bool) []struct{ IsResolved bool } {
		nodes := make([]struct{ IsResolved bool }, len(resolved))
		for i, r := range resolved {
			nodes[i] = struct{ IsResolved bool }{IsResolved: r}
		}
		return nodes
	}

	tests := []struct {
		name  string
		nodes []struct{ IsResolved bool }
		want  int
	}{
		{
			name:  "no threads",
			nodes: newNodes(),
			want:  0,
		},
		{
			name:  "all resolved",
			nodes: newNodes(true, true, true),
			want:  0,
		},
		{
			name:  "some unresolved",
			nodes: newNodes(false, true, false),
			want:  2,
		},
		{
			name:  "all unresolved",
			nodes: newNodes(false, false),
			want:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pr := PullRequestData{
				ReviewThreads: ReviewThreadsWithResolution{Nodes: tt.nodes},
			}
			require.Equal(t, tt.want, pr.UnresolvedThreadsCount())
		})
	}
}

func TestReviewThreadWithComments_Fields(t *testing.T) {
	// FetchReviewThreads itself talks to GitHub's GraphQL API and isn't
	// unit-tested here (matching FetchPullRequest/FetchPullRequests, which
	// have no direct unit tests either); the queue-building logic that
	// excludes IsResolved: true threads is exercised where it lives, in the
	// prview package. This just pins down the struct's new fields.
	thread := ReviewThreadWithComments{
		Id:               "thread-1",
		IsResolved:       true,
		ViewerCanReply:   true,
		ViewerCanResolve: false,
		Path:             "main.go",
		Line:             42,
		Comments: ReviewComments{
			Nodes: []ReviewComment{
				{Body: "looks good", DiffHunk: "@@ -1,2 +1,2 @@\n-old\n+new"},
			},
		},
	}

	require.True(t, thread.IsResolved)
	require.True(t, thread.ViewerCanReply)
	require.False(t, thread.ViewerCanResolve)
	require.Equal(t, "main.go", thread.Path)
	require.Equal(t, "@@ -1,2 +1,2 @@\n-old\n+new", thread.Comments.Nodes[0].DiffHunk)
}

func TestSetClient(t *testing.T) {
	// Save original state
	originalClient := client
	originalCachedClient := cachedClient
	defer func() {
		client = originalClient
		cachedClient = originalCachedClient
	}()

	t.Run("sets both client and cachedClient", func(t *testing.T) {
		client = nil
		cachedClient = nil

		// SetClient with nil should set both to nil
		SetClient(nil)
		require.Nil(t, client)
		require.True(t, IsEnrichmentCacheCleared())
	})
}
