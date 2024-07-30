package dbusers_test

import (
	"testing"
	"tgssn/internal/model/db"
	"tgssn/tests/suite"

	"github.com/stretchr/testify/assert"
)

func TestWorkersStorage(t *testing.T) {
	ctx, st := suite.New(t)

	storage := db.NewUserStorage(st.Db)

	t.Run("worker exists", func(t *testing.T) {
		userID := int64(1)
		exist, err := storage.CheckIfWorkerExist(ctx, userID)
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("worker does not exist", func(t *testing.T) {
		userID := int64(100)
		exist, err := storage.CheckIfWorkerExist(ctx, userID)
		assert.NoError(t, err)
		assert.False(t, exist)
	})

	t.Run("add worker", func(t *testing.T) {
		userID := int64(2)
		exist, err := storage.CheckIfWorkerExistAndAdd(ctx, userID)
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("try to add existed worker", func(t *testing.T) {
		userID := int64(1)
		exist, err := storage.CheckIfWorkerExistAndAdd(ctx, userID)
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("get all workers", func(t *testing.T) {
		workers, err := storage.GetAllWorkers(ctx, int64(1))
		assert.NoError(t, err)
		assert.NotEmpty(t, workers)
	})

	t.Run("change worker status", func(t *testing.T) {
		userID := int64(1)
		status := true
		err := storage.ChangeWorkerStatus(ctx, userID, status)
		assert.NoError(t, err)
	})

	t.Run("create ticket", func(t *testing.T) {
		userID := int64(1)
		err := storage.CreateTicket(ctx, userID)
		assert.NoError(t, err)
	})

	t.Run("update ticket status", func(t *testing.T) {
		userID := int64(1)
		status := "good"
		err := storage.UpdateTicketStatus(ctx, userID, status)
		assert.NoError(t, err)
	})

	t.Run("count workers statistic", func(t *testing.T) {
		userID := int64(1)
		goods, bads, err := storage.CountWorkersStatistic(ctx, userID)
		assert.NoError(t, err)
		assert.NotZero(t, goods)
		assert.Zero(t, bads)
	})
}
