package durablemetrics

const (
	prefix = "metric_"
)

func generateRedisKey(originalKey string) string {
	return prefix + originalKey
}
