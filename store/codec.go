package store

import (
	"encoding/json"
	"fmt"
	"go.etcd.io/bbolt"
	"reflect"
)

func encode(value any) ([]byte, error) { return json.Marshal(value) }

func decode(data []byte, target any) error {
	if len(data) == 0 {
		return fmt.Errorf("empty value")
	}
	return json.Unmarshal(data, target)
}

func putJSON(tx *bbolt.Tx, bucket, key []byte, value any) error {
	data, err := encode(value)
	if err != nil {
		return err
	}
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	return b.Put(key, data)
}

func getJSON(tx *bbolt.Tx, bucket, key []byte, target any) error {
	b := tx.Bucket(bucket)
	if b == nil {
		return fmt.Errorf("bucket %s missing", bucket)
	}
	value := b.Get(key)
	if value == nil {
		return fmt.Errorf("key %s missing", key)
	}
	return decode(append([]byte(nil), value...), target)
}

func cloneValue[T any](source T) (T, error) {
	var result T
	data, err := encode(source)
	if err != nil {
		return result, err
	}
	err = decode(data, &result)
	return result, err
}

func sameType(a, b any) bool { return reflect.TypeOf(a) == reflect.TypeOf(b) }
