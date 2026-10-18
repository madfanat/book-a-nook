-- +goose Up
CREATE INDEX slots_resource_id_id ON slots (resource_id, id);

-- +goose Down
DROP INDEX slots_resource_id_id;
