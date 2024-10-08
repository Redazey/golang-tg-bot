package dbusers_test

import (
	"testing"

	"tgseller/internal/model/db"
	"tgseller/tests/suite"

	"github.com/stretchr/testify/assert"
)

func TestUserStorage(t *testing.T) {
	ctx, st := suite.New(t)

	storage := db.NewUserStorage(st.Db)

	t.Run("TestInsertUser", func(t *testing.T) {
		err := storage.InsertUser(ctx, 123456)
		assert.NoError(t, err)
	})

	t.Run("TestCheckIfUserExist", func(t *testing.T) {
		exist, err := storage.CheckIfUserExist(ctx, 123456)
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("TestCheckIfUserExistAndAdd", func(t *testing.T) {
		exist, err := storage.CheckIfUserExistAndAdd(ctx, 654321)
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	var currLimit float64
	t.Run("TestGetUserLimit", func(t *testing.T) {
		limit, err := storage.GetUserLimit(ctx, 123456)
		assert.NoError(t, err)
		assert.IsType(t, currLimit, limit)
		currLimit = limit
	})

	t.Run("TestAddUserLimit", func(t *testing.T) {
		err := storage.AddUserLimit(ctx, 123456, 100)
		assert.NoError(t, err)

		limit, err := storage.GetUserLimit(ctx, 123456)
		assert.NoError(t, err)
		assert.Equal(t, float64(currLimit+100), limit)
	})
}
