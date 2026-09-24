package dnsfmt

// DiffLines compares two slices of lines -- typically the output of two
// NormalizeZoneFile calls -- and returns a unified-diff-style listing:
// a line present in both is prefixed with a space, a line only in a is
// prefixed with "-", and a line only in b is prefixed with "+". Because
// the inputs are expected to already be normalized, a non-space-prefixed
// line in the result reflects a real difference in the records rather
// than formatting noise like case, whitespace, or TTL units.
//
// The comparison is a standard longest-common-subsequence diff, so
// lines that moved without changing are matched up rather than shown as
// a delete-then-add pair whenever a shared ordering exists.
func DiffLines(a, b []string) []string {
	n, m := len(a), len(b)

	// dp[i][j] holds the LCS length of a[i:] and b[j:].
	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			switch {
			case a[i] == b[j]:
				dp[i][j] = dp[i+1][j+1] + 1
			case dp[i+1][j] >= dp[i][j+1]:
				dp[i][j] = dp[i+1][j]
			default:
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	out := make([]string, 0, n+m)
	i, j := 0, 0
	for i < n && j < m {
		switch {
		case a[i] == b[j]:
			out = append(out, " "+a[i])
			i++
			j++
		case dp[i+1][j] >= dp[i][j+1]:
			out = append(out, "-"+a[i])
			i++
		default:
			out = append(out, "+"+b[j])
			j++
		}
	}
	for ; i < n; i++ {
		out = append(out, "-"+a[i])
	}
	for ; j < m; j++ {
		out = append(out, "+"+b[j])
	}
	return out
}
