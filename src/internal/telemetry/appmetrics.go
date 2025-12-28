package telemetry

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

const (
	queryUsersTotal     = "SELECT COUNT(*) FROM users"
	queryUsersNew24h    = "SELECT COUNT(*) FROM users WHERE created_at >= NOW() - INTERVAL '24 hours'"
	querySnippetsTotal  = "SELECT COUNT(*) FROM snippets"
	querySnippetsNew24h = "SELECT COUNT(*) FROM snippets WHERE created_at >= NOW() - INTERVAL '24 hours'"
)

func InitAppMetrics(serviceName string, db *pgxpool.Pool, redisClient *redis.Client, sessionPrefix string) {
	meter := otel.Meter(serviceName)

	usersTotal, err := meter.Int64ObservableGauge(
		"sniply_users_total",
		metric.WithDescription("Total users"),
	)
	if err != nil {
		return
	}
	usersNew, err := meter.Int64ObservableGauge(
		"sniply_users_new_24h",
		metric.WithDescription("New users in the last 24 hours"),
	)
	if err != nil {
		return
	}
	snippetsTotal, err := meter.Int64ObservableGauge(
		"sniply_snippets_total",
		metric.WithDescription("Total snippets"),
	)
	if err != nil {
		return
	}
	snippetsNew, err := meter.Int64ObservableGauge(
		"sniply_snippets_new_24h",
		metric.WithDescription("New snippets in the last 24 hours"),
	)
	if err != nil {
		return
	}
	sessionsActive, err := meter.Int64ObservableGauge(
		"sniply_sessions_active",
		metric.WithDescription("Active sessions (Redis keys)"),
	)
	if err != nil {
		return
	}

	_, _ = meter.RegisterCallback(func(ctx context.Context, o metric.Observer) error {
		var usersTotalVal int64
		var usersNewVal int64
		var snippetsTotalVal int64
		var snippetsNewVal int64

		if db != nil {
			dbCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			usersTotalVal = queryCount(dbCtx, db, queryUsersTotal)
			usersNewVal = queryCount(dbCtx, db, queryUsersNew24h)
			snippetsTotalVal = queryCount(dbCtx, db, querySnippetsTotal)
			snippetsNewVal = queryCount(dbCtx, db, querySnippetsNew24h)
			cancel()
		}

		o.ObserveInt64(usersTotal, usersTotalVal)
		o.ObserveInt64(usersNew, usersNewVal)
		o.ObserveInt64(snippetsTotal, snippetsTotalVal)
		o.ObserveInt64(snippetsNew, snippetsNewVal)

		var sessionsVal int64
		if redisClient != nil && sessionPrefix != "" {
			redisCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			sessionsVal = countRedisKeys(redisCtx, redisClient, sessionPrefix+"*")
			cancel()
		}
		o.ObserveInt64(sessionsActive, sessionsVal)
		return nil
	}, usersTotal, usersNew, snippetsTotal, snippetsNew, sessionsActive)
}

func queryCount(ctx context.Context, db *pgxpool.Pool, sql string) int64 {
	var count int64
	if err := db.QueryRow(ctx, sql).Scan(&count); err != nil {
		return 0
	}
	return count
}

func countRedisKeys(ctx context.Context, client *redis.Client, pattern string) int64 {
	var total int64
	iter := client.Scan(ctx, 0, pattern, 1000).Iterator()
	for iter.Next(ctx) {
		total++
	}
	if err := iter.Err(); err != nil {
		return 0
	}
	return total
}
