package ember

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JsonColumn[T any] struct {
	Data  T
	Valid bool
}

func (j *JsonColumn[T]) Scan(src any) error {
	var res T

	if src == nil {
		j.Data = res
		j.Valid = false
		return nil
	}

	switch value := src.(type) {
	case string:
		err := json.Unmarshal([]byte(value), &j.Data)
		if err != nil {
			return fmt.Errorf("jsoncolumn: failed to unmarshal data: %w", err)
		}

		j.Valid = true
	case []byte:
		err := json.Unmarshal(value, &j.Data)
		if err != nil {
			return fmt.Errorf("jsoncolumn: failed to unmarshal data: %w", err)
		}

		j.Valid = true
	default:
		return fmt.Errorf("jsoncolumn: unsupported src type %T", src)
	}

	return nil
}

func (j *JsonColumn[T]) Value() (driver.Value, error) {
	raw, err := json.Marshal(j.Data)
	return raw, err
}

type KVStore map[string]string

func (kv KVStore) Serialize() (string, error) {
	b, err := json.Marshal(kv)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func Deserialize(data string) (KVStore, error) {
	kv := make(KVStore)
	if data == "" {
		return kv, nil
	}

	err := json.Unmarshal([]byte(data), &kv)
	if err != nil {
		return nil, err
	}

	return kv, nil
}

func (kv KVStore) Value() (driver.Value, error) {
	return kv.Serialize()
}

func (kv *KVStore) Scan(src any) error {
	if src == nil {
		*kv = make(KVStore)
		return nil
	}

	switch value := src.(type) {
	case string:
		r, err := Deserialize(value)
		if err != nil {
			return fmt.Errorf("kvstore: failed to deserialize store: %w", err)
		}

		*kv = r
	case []byte:
		r, err := Deserialize(string(value))
		if err != nil {
			return fmt.Errorf("kvstore: failed to deserialize store: %w", err)
		}

		*kv = r
	default:
		return fmt.Errorf("kvstore: unsupported src type %T", src)
	}

	return nil
}
