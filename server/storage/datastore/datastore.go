package datastore

import "appengine"
import "appengine/datastore"

type DataStore struct {
	Context appengine.Context
}

type ValueType struct {
	Value []byte
}

func (ds *DataStore) Save(kind string, key, value []byte) error {
	k := datastore.NewKey(ds.Context, kind, string(key), 0, nil)
	v := &ValueType{value}
	_, err := datastore.Put(ds.Context, k, v)
	return err
}

func (ds *DataStore) Load(kind string, key []byte) ([]byte, error) {
	k := datastore.NewKey(ds.Context, kind, string(key), 0, nil)
	v := &ValueType{}
	err := datastore.Get(ds.Context, k, v)
	return v.Value, err
}
