package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
)

func (h *Handler) metricsSnapshot(ctx context.Context) (map[string]any, time.Time, error) {
	if h.db == nil {
		return nil, time.Time{}, fmt.Errorf("metrics database is not configured")
	}
	var sampledAt time.Time
	var cpu float64
	var ramUsed, ramTotal, diskUsed, diskTotal, databaseBytes, users, rooms, online int64
	var prometheusJSON string
	err := h.db.QueryRow(ctx, "SELECT sampled_at, COALESCE(cpu_percent, 0)::double precision, COALESCE(ram_used_bytes, 0)::bigint, COALESCE(ram_total_bytes, 0)::bigint, COALESCE(disk_used_bytes, 0)::bigint, COALESCE(disk_total_bytes, 0)::bigint, COALESCE(database_bytes, 0)::bigint, COALESCE(users_total, 0)::bigint, COALESCE(rooms_total, 0)::bigint, COALESCE(online_users, 0)::bigint, COALESCE(prometheus, '{}'::jsonb)::text FROM system_metrics ORDER BY sampled_at DESC LIMIT 1").Scan(&sampledAt, &cpu, &ramUsed, &ramTotal, &diskUsed, &diskTotal, &databaseBytes, &users, &rooms, &online, &prometheusJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, time.Time{}, nil
	}
	if err != nil {
		return nil, time.Time{}, err
	}

	var prometheus any = map[string]any{}
	if prometheusJSON != "" {
		if err := json.Unmarshal([]byte(prometheusJSON), &prometheus); err != nil {
			prometheus = map[string]any{"parse_error": "invalid Prometheus JSON"}
		}
	}
	return map[string]any{
		"sampled_at": sampledAt, "cpu_percent": cpu,
		"ram_used_bytes": ramUsed, "ram_total_bytes": ramTotal,
		"disk_used_bytes": diskUsed, "disk_total_bytes": diskTotal,
		"database_bytes": databaseBytes, "users_total": users,
		"rooms_total": rooms, "online_users": online,
		"prometheus": prometheus,
	}, sampledAt, nil
}

// MetricsCurrent returns the newest persisted metrics sample.
func (h *Handler) MetricsCurrent(w http.ResponseWriter, r *http.Request) {
	snapshot, _, err := h.metricsSnapshot(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("reading current metrics failed")
		write(w, http.StatusServiceUnavailable, envelope{OK: false, Error: &apiError{Code: "metrics_unavailable", Message: "Unable to read system metrics"}})
		return
	}
	write(w, http.StatusOK, envelope{OK: true, Data: snapshot})
}

// MetricsStream streams new collector samples to authenticated dashboard clients.
func (h *Handler) MetricsStream(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		write(w, http.StatusInternalServerError, envelope{OK: false, Error: &apiError{Code: "stream_unsupported", Message: "Streaming is not supported by this server"}})
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache, no-transform")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, ": connected\n\n")
	flusher.Flush()

	var lastSample time.Time
	send := func(snapshot map[string]any, sampledAt time.Time) bool {
		if snapshot == nil || !sampledAt.After(lastSample) {
			return true
		}
		payload, err := json.Marshal(snapshot)
		if err != nil {
			return false
		}
		if _, err = fmt.Fprintf(w, "id: %d\nevent: metrics\ndata: %s\n\n", sampledAt.UnixNano(), payload); err != nil {
			return false
		}
		lastSample = sampledAt
		flusher.Flush()
		return true
	}

	snapshot, sampledAt, err := h.metricsSnapshot(r.Context())
	if err != nil {
		h.logger.Warn().Err(err).Msg("initial metrics stream snapshot failed")
		_, _ = fmt.Fprint(w, "event: error\ndata: {\"message\":\"metrics unavailable\"}\n\n")
		flusher.Flush()
		return
	}
	_ = send(snapshot, sampledAt)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			snapshot, sampledAt, err := h.metricsSnapshot(r.Context())
			if err != nil {
				h.logger.Warn().Err(err).Msg("metrics stream query failed")
				_, _ = fmt.Fprint(w, "event: error\ndata: {\"message\":\"metrics query failed\"}\n\n")
				flusher.Flush()
				return
			}
			if !send(snapshot, sampledAt) {
				return
			}
		case <-heartbeat.C:
			_, _ = fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}
