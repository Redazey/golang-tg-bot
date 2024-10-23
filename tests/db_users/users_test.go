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
}
