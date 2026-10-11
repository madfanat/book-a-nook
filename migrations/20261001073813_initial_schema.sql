-- +goose Up

CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE resources (
    id text PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id text PRIMARY KEY,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE slots (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource_id text NOT NULL,
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT slots_resource_fk
        FOREIGN KEY (resource_id) REFERENCES resources(id),

    CONSTRAINT slots_valid_time_range
        CHECK (ends_at > starts_at),

    CONSTRAINT slots_no_overlap
        EXCLUDE USING gist (
            resource_id WITH =,
            tstzrange(starts_at, ends_at, '[)') WITH &&
        )
);

CREATE TABLE bookings (
    id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slot_id integer NOT NULL,
    user_id text NOT NULL,
    status text NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),

    CONSTRAINT bookings_slot_fk
        FOREIGN KEY (slot_id) REFERENCES slots(id),

    CONSTRAINT bookings_user_fk
        FOREIGN KEY (user_id) REFERENCES users(id),

    CONSTRAINT bookings_status
        CHECK (status IN ('active', 'cancelled'))
);

CREATE UNIQUE INDEX bookings_one_active_per_slot
    ON bookings(slot_id)
    WHERE status = 'active';

-- +goose Down
DROP TABLE bookings;
DROP TABLE slots;
DROP TABLE users;
DROP TABLE resources;
