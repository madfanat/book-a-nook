-- +goose Up

ALTER TABLE bookings
    ADD COLUMN request_key text,
    ADD COLUMN request_hash text;

UPDATE bookings
SET request_key = '__legacy_booking__:' || id::text,
    request_hash = encode(
        sha256(convert_to('slot_id:' || slot_id::text, 'UTF8')),
        'hex'
    );

ALTER TABLE bookings
    ALTER COLUMN request_key SET NOT NULL,
    ALTER COLUMN request_hash SET NOT NULL,
    ADD CONSTRAINT bookings_user_request_key_unique
        UNIQUE (user_id, request_key);

-- +goose Down

ALTER TABLE bookings
    DROP CONSTRAINT bookings_user_request_key_unique,
    DROP COLUMN request_hash,
    DROP COLUMN request_key;
