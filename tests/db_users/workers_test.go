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

	t.Run("add worker", func(t *testing.T) {
		userID := int64(1)
		exist, err := storage.CheckIfWorkerExistAndAdd(ctx, userID, "sigma")
		assert.NoError(t, err)
		assert.True(t, exist)
	})

	t.Run("try to add existed worker", func(t *testing.T) {
		userID := int64(1)
		exist, err := storage.CheckIfWorkerExistAndAdd(ctx, userID, "sigma")
		assert.NoError(t, err)
		assert.True(t, exist)
	})

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

	t.Run("get all workers", func(t *testing.T) {
		workers, err := storage.GetAllWorkers(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, workers)
	})

	t.Run("create ticket", func(t *testing.T) {
		workerID := int64(1)
		buyerID := int64(123456)
		ctgID := int64(1)
		_, err := storage.CreateTicket(ctx, workerID, buyerID, ctgID)
		assert.NoError(t, err)
	})

	t.Run("create busy ticket", func(t *testing.T) {
		workerID := int64(1)
		buyerID := int64(123456)
		ctgID := int64(1)
		succsessful, err := storage.CreateTicket(ctx, workerID, buyerID, ctgID)
		assert.NoError(t, err)
		assert.False(t, succsessful)
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
