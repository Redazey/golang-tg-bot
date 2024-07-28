package dbusers_test

import (
	"testing"

	"tgssn/internal/model/bottypes"
	"tgssn/internal/model/db"
	"tgssn/tests/suite"

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

	t.Run("TestInsertUserDataRecord", func(t *testing.T) {
		isOverLimit, err := storage.InsertUserDataRecord(ctx, 123456, bottypes.UserDataRecord{
			Category: "test_category",
			Sum:      100,
		})
		assert.NoError(t, err)
		assert.False(t, isOverLimit)
	})

	t.Run("TestInsertUserDataRecordOverLimit", func(t *testing.T) {
		isOverLimit, err := storage.InsertUserDataRecord(ctx, 123456, bottypes.UserDataRecord{
			Category: "test_category",
			Sum:      1000000,
		})
		assert.NoError(t, err)
		assert.False(t, isOverLimit)
	})

	t.Run("TestGetUserLimit", func(t *testing.T) {
		limit, err := storage.GetUserLimit(ctx, 123456)
		assert.NoError(t, err)
		assert.Equal(t, float64(0), limit)
	})

	t.Run("TestAddUserLimit", func(t *testing.T) {
		err := storage.AddUserLimit(ctx, 123456, 100)
		assert.NoError(t, err)

		limit, err := storage.GetUserLimit(ctx, 123456)
		assert.NoError(t, err)
		assert.Equal(t, float64(100), limit)
	})
}
