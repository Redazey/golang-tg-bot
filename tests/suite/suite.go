package suite

import (
	"context"
	"testing"
	"tgssn/config"
	"tgssn/pkg/db"
	"tgssn/pkg/logger"
	"time"

	"github.com/jmoiron/sqlx"
)

type Suite struct {
	*testing.T
	Env *config.Enviroment
	Db  *sqlx.DB
}

// New creates new test suite.
func New(t *testing.T) (context.Context, *Suite) {
	t.Helper() // Функция будет восприниматься как вспомогательная для тестов

	// Читаем конфиг из файла
	env, err := config.NewEnv("../../.env")
	if err != nil {
		t.Fatalf("ошибка при инициализации файла конфигурации: %s", err)
	}

	logger.Init(env.LoggerLevel, "../../logs/golog")

	err = db.Init(env.DB.DBUser, env.DB.DBPassword, env.DB.DBHost, env.DB.DBName)
	if err != nil {
		t.Fatalf("ошибка при инициализации БД: %s", err)
	}

	// Основной родительский контекст
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)

	// Когда тесты пройдут, закрываем контекст
	t.Cleanup(func() {
		t.Helper()
		cancelCtx()
	})

	return ctx, &Suite{
		T:   t,
		Env: env,
		Db:  db.Conn,
	}
}
