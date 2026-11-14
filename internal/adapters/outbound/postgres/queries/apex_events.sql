-- name: GetApexRestRoute :one
SELECT *
FROM salesforce.apex_rest_route
WHERE url_mapping = @url_mapping::TEXT
  AND upper(http_method) = upper(@http_method::TEXT)
  AND is_active = TRUE
LIMIT 1;

-- name: CreatePlatformEvent :one
INSERT INTO salesforce.platform_event (
    channel_name,
    event_type,
    payload
)
VALUES (
    @channel_name::TEXT,
    @event_type::TEXT,
    @payload::TEXT
)
RETURNING *;

-- name: ListPlatformEventsSince :many
SELECT *
FROM salesforce.platform_event
WHERE channel_name = @channel_name::TEXT
  AND replay_id > @after_replay_id::BIGINT
ORDER BY replay_id ASC;

-- name: UpsertCometDClient :one
INSERT INTO salesforce.cometd_client (
    client_id,
    is_connected,
    last_seen_at
)
VALUES (
    @client_id::TEXT,
    TRUE,
    CURRENT_TIMESTAMP
)
ON CONFLICT (client_id) DO UPDATE
SET is_connected = TRUE,
    last_seen_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetCometDClient :one
SELECT *
FROM salesforce.cometd_client
WHERE client_id = @client_id::TEXT
LIMIT 1;

-- name: DisconnectCometDClient :exec
UPDATE salesforce.cometd_client
SET is_connected = FALSE,
    last_seen_at = CURRENT_TIMESTAMP
WHERE client_id = @client_id::TEXT;

-- name: UpsertCometDSubscription :one
INSERT INTO salesforce.cometd_subscription (
    client_id,
    channel_name,
    last_replay_id
)
VALUES (
    @client_id::TEXT,
    @channel_name::TEXT,
    @last_replay_id::BIGINT
)
ON CONFLICT (client_id, channel_name) DO UPDATE
SET last_replay_id = EXCLUDED.last_replay_id
RETURNING *;

-- name: DeleteCometDSubscription :exec
DELETE FROM salesforce.cometd_subscription
WHERE client_id = @client_id::TEXT
  AND channel_name = @channel_name::TEXT;

-- name: ListCometDSubscriptionsByClient :many
SELECT *
FROM salesforce.cometd_subscription
WHERE client_id = @client_id::TEXT
ORDER BY id ASC;

-- name: UpdateCometDSubscriptionReplayID :exec
UPDATE salesforce.cometd_subscription
SET last_replay_id = @last_replay_id::BIGINT
WHERE id = @id::BIGINT;
