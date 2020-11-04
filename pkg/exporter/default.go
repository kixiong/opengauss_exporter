// Copyright © 2020 Bin Liu <bin.liu@enmotech.com>

package exporter

// var (
// 	ogVersionName = "OG_VERSION"
// )

var (
	pgLock = &QueryInstance{
		Name: "pg_lock",
		Desc: "OpenGauss lock distribution by mode",
		Queries: []*Query{
			{
				SupportedVersions: ">=0.0.0",
				SQL: `SELECT
  pg_database.datname,
  tmp.mode,
  COALESCE(count,0) as count
FROM
    (
      VALUES ('accesssharelock'),
             ('rowsharelock'),
             ('rowexclusivelock'),
             ('shareupdateexclusivelock'),
             ('sharelock'),
             ('sharerowexclusivelock'),
             ('exclusivelock'),
             ('accessexclusivelock')
    ) AS tmp(mode) CROSS JOIN pg_database
LEFT JOIN
  (SELECT database, lower(mode) AS mode,count(*) AS count
  FROM pg_locks WHERE database IS NOT NULL
  GROUP BY database, lower(mode)
) AS tmp2
ON tmp.mode=tmp2.mode and pg_database.oid = tmp2.database ORDER BY 1`,
			},
		},
		Metrics: []*Column{
			{Name: "datname", Desc: "Name of this database", Usage: LABEL},
			{Name: "mode", Desc: "Type of Lock", Usage: LABEL},
			{Name: "count", Desc: "Number of locks", Usage: GAUGE},
		},
	}
	pgStatReplication = &QueryInstance{
		Name: "pg_stat_replication",
		Desc: "",
		Queries: []*Query{
			{
				Name: "pg_stat_replication",
				SQL: `SELECT *,
  (case pg_is_in_recovery() when 't' then null else pg_current_xlog_location() end) AS pg_current_xlog_location,
  (case pg_is_in_recovery() when 't' then null else pg_xlog_location_diff(pg_current_xlog_location(), receiver_replay_location)::float end) AS pg_xlog_location_diff
FROM pg_stat_replication`,
				SupportedVersions: ">=1.0.0",
			},
		},
		Metrics: []*Column{
			{Name: "pid", Usage: DISCARD, Desc: "Process ID of a WAL sender process"},
			{Name: "usesysid", Usage: DISCARD, Desc: "OID of the user logged into this WAL sender process"},
			{Name: "usename", Usage: DISCARD, Desc: "Name of the user logged into this WAL sender process"},
			{Name: "application_name", Usage: LABEL, Desc: "Name of the application that is connected to this WAL sender"},
			{Name: "client_addr", Usage: LABEL, Desc: "IP address of the client connected to this WAL sender. If this field is null, it indicates that the client is connected via a Unix socket on the server machine."},
			{Name: "client_hostname", Usage: DISCARD, Desc: "Host name of the connected client, as reported by a reverse DNS lookup of client_addr. This field will only be non-null for IP connections, and only when log_hostname is enabled."},
			{Name: "client_port", Usage: DISCARD, Desc: "TCP port number that the client is using for communication with this WAL sender, or -1 if a Unix socket is used"},
			{Name: "backend_start", Usage: DISCARD, Desc: "with time zone      Time when this process was started, i.e., when the client connected to this WAL sender"},
			{Name: "state", Usage: LABEL, Desc: "Current WAL sender state"},
			{Name: "sender_sent_location", Usage: GAUGE, Desc: "Last transaction log position sent on this connection"},
			{Name: "receiver_write_location", Usage: GAUGE, Desc: "Last transaction log position written to disk by this standby server"},
			{Name: "receiver_flush_location", Usage: GAUGE, Desc: "Last transaction log position flushed to disk by this standby server"},
			{Name: "receiver_replay_location", Usage: GAUGE, Desc: "Last transaction log position replayed into the database on this standby server"},
			{Name: "sync_priority", Usage: DISCARD, Desc: "Priority of this standby server for being chosen as the synchronous standby"},
			{Name: "sync_state", Usage: GAUGE, Desc: "Synchronous state of this standby server"},
			{Name: "pg_current_xlog_location", Usage: DISCARD, Desc: "pg_current_xlog_location"},
			{Name: "pg_xlog_location_diff", Usage: GAUGE, Desc: "Lag in bytes between master and slave"},
		},
	}
	pgStatActivity = &QueryInstance{
		Name: "pg_stat_activity",
		Desc: "",
		Queries: []*Query{
			{
				SQL: `SELECT
  pg_database.datname,
  tmp.state,
  COALESCE(count,0) as count,
  COALESCE(max_tx_duration,0) as max_tx_duration
FROM
  (
    VALUES ('active'),
         ('idle'),
         ('idle in transaction'),
         ('idle in transaction (aborted)'),
         ('fastpath function call'),
         ('disabled')
  ) AS tmp(state) CROSS JOIN pg_database
LEFT JOIN
(
  SELECT
    datname,
    state,
    count(*) AS count,
    MAX(EXTRACT(EPOCH FROM now() - xact_start))::float AS max_tx_duration
  FROM pg_stat_activity GROUP BY datname,state) AS tmp2
  ON tmp.state = tmp2.state AND pg_database.datname = tmp2.datname`,
				SupportedVersions: ">=1.0.0",
			},
		},
		Metrics: []*Column{
			{Name: "datname", Usage: LABEL, Desc: "Name of this database"},
			{Name: "state", Usage: LABEL, Desc: "connection state"},
			{Name: "count", Usage: GAUGE, Desc: "number of connections in this state"},
			{Name: "max_tx_duration", Usage: GAUGE, Desc: "max duration in seconds any active transaction has been running"},
		},
	}
	pgDatabase = &QueryInstance{
		Name: "pg_database",
		Desc: "",
		Queries: []*Query{
			{
				SQL:               `SELECT pg_database.datname, pg_database_size(pg_database.datname) as size_bytes FROM pg_database`,
				SupportedVersions: ">=0.0.0",
			},
		},
		Metrics: []*Column{
			{Name: "datname", Usage: LABEL, Desc: "Name of this database"},
			{Name: "size_bytes", Usage: GAUGE, Desc: "Disk space used by the database"},
		},
	}
	pgBgWriter = &QueryInstance{
		Name: "pg_bgwriter",
		Desc: "",
		Queries: []*Query{
			{
				SQL: `SELECT checkpoints_timed,
    checkpoints_req,
    checkpoint_write_time,
    checkpoint_sync_time,
    buffers_checkpoint,
    buffers_clean,
    buffers_backend,
    maxwritten_clean,
    buffers_backend_fsync,
    buffers_alloc,
    stats_reset
FROM pg_stat_bgwriter;;`,
				SupportedVersions: ">=0.0.0",
			},
		},
		Metrics: []*Column{
			{Name: "checkpoints_timed", Usage: COUNTER, Desc: "scheduled checkpoints that have been performed"},
			{Name: "checkpoints_req", Usage: COUNTER, Desc: "requested checkpoints that have been performed"},
			{Name: "checkpoint_write_time", Usage: COUNTER, Desc: "time spending on writing files to disk, in µs"},
			{Name: "checkpoint_sync_time", Usage: COUNTER, Desc: "time spending on syncing files to disk, in µs"},
			{Name: "buffers_checkpoint", Usage: COUNTER, Desc: "buffers written during checkpoints"},
			{Name: "buffers_clean", Usage: COUNTER, Desc: "buffers written by the background writer"},
			{Name: "buffers_backend", Usage: COUNTER, Desc: "buffers written directly by a backend"},
			{Name: "maxwritten_clean", Usage: COUNTER, Desc: "times that bgwriter stopped a cleaning scan"},
			{Name: "buffers_backend_fsync", Usage: COUNTER, Desc: "times a backend had to execute its own fsync"},
			{Name: "buffers_alloc", Usage: COUNTER, Desc: "buffers allocated"},
			{Name: "stats_reset", Usage: COUNTER, Desc: "time when statistics were last reset"},
		},
	}
)

var (
	defaultMonList = map[string]*QueryInstance{
		"pg_lock":             pgLock,
		"pg_stat_replication": pgStatReplication,
		"pg_stat_activity":    pgStatActivity,
		"pg_database":         pgDatabase,
		"pg_bgwriter":         pgBgWriter,
	}
)
