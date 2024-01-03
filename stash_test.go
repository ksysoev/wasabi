package wasabi

import (
	"reflect"
	"testing"
)

func TestNewStashStore(t *testing.T) {
	store := NewStashStore()

	if store == nil {
		t.Errorf("NewStashStore() returned nil")
	}

	if reflect.TypeOf(store) != reflect.TypeOf(&StashStore{}) {
		t.Errorf("NewStashStore() returned wrong type, got %T, want %T", store, &StashStore{})
	}

	if len(store.data) != 0 {
		t.Errorf("NewStashStore() returned store with non-empty data map")
	}
}

func TestStashStore_Set(t *testing.T) {
	store := NewStashStore()

	key := "testKey"
	value := "testValue"

	store.Set(key, value)

	if store.data[key] != value {
		t.Errorf("Set() failed to set the value for the key, got %v, want %v", store.data[key], value)
	}
}
func TestStashStore_Get(t *testing.T) {
	store := NewStashStore()

	key := "testKey"
	value := "testValue"
	store.data[key] = value

	result := store.Get(key)

	if result != value {
		t.Errorf("Get() failed to retrieve the correct value, got %v, want %v", result, value)
	}
}
func TestStashStore_Delete(t *testing.T) {
	store := NewStashStore()

	key := "testKey"
	value := "testValue"
	store.data[key] = value

	store.Delete(key)

	if _, ok := store.data[key]; ok {
		t.Errorf("Delete() failed to delete the key, key still exists in the store")
	}
}
