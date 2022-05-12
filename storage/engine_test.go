package storage

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/nuts-foundation/nuts-node/core"
	"github.com/nuts-foundation/nuts-node/test/io"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_New(t *testing.T) {
	assert.NotNil(t, New())
}

func Test_engine_Name(t *testing.T) {
	assert.Equal(t, "Storage", engine{}.Name())
}

func Test_engine_lifecycle(t *testing.T) {
	sut := New()
	err := sut.Configure(core.ServerConfig{Datadir: io.TestDirectory(t)})
	if !assert.NoError(t, err) {
		return
	}
	err = sut.Start()
	if !assert.NoError(t, err) {
		return
	}
	// Get a KV store so there's something to shut down
	_, err = sut.GetKVStore("test", "store")
	if !assert.NoError(t, err) {
		return
	}
	err = sut.Shutdown()
	if !assert.NoError(t, err) {
		return
	}
}

func Test_engine_GetKVStore(t *testing.T) {
	sut := New()
	t.Run("key is empty", func(t *testing.T) {
		store, err := sut.GetKVStore("", "store")
		assert.Nil(t, store)
		assert.EqualError(t, err, "invalid engine key")
	})
	t.Run("store is empty", func(t *testing.T) {
		store, err := sut.GetKVStore("engine", "")
		assert.Nil(t, store)
		assert.EqualError(t, err, "invalid store name")
	})
}

func Test_engine_Shutdown(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store := NewMockKVStore(ctrl)
		store.EXPECT().Close()

		sut := New().(*engine)
		sut.stores["1"] = store

		err := sut.Shutdown()

		assert.NoError(t, err)
	})
	t.Run("error while closing store results in error, but all stores are closed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		store1 := NewMockKVStore(ctrl)
		store1.EXPECT().Close().Return(errors.New("failed"))
		store2 := NewMockKVStore(ctrl)
		store2.EXPECT().Close().Return(errors.New("failed"))

		sut := New().(*engine)
		sut.stores["1"] = store1
		sut.stores["2"] = store2

		err := sut.Shutdown()

		assert.EqualError(t, err, "one or more stores failed to close")
	})
}
