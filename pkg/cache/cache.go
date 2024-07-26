package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	Ctx         = context.Background()
	Rdb         *redis.Client
	CacheEXTime time.Duration
)

func Init(Addr string, Password string, DB int, CacheEx time.Duration) error {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     Addr,
		Password: Password,
		DB:       DB,
	})

	err := Rdb.Ping(Ctx).Err()
	if err != nil {
		return err
	}

	CacheEXTime = CacheEx

	return nil
}

/*
функция для проверки существования таблицы в кэше

принимает:

	table - имя таблицы

возвращает:

	bool - true, если таблица существует, иначе false
	error - ошибка, если возникла
*/
func IsExistInCache(hashKey string) (bool, error) {
	exists, err := Rdb.Exists(Ctx, hashKey).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

/*
функция для записи данных в кэш как string, принимает любые данные на вход
*/
func SaveCache(hashKey string, data interface{}) error {
	Rdb.Set(Ctx, hashKey, data, time.Minute*CacheEXTime).Err()

	return nil
}

/*
Функция для чтения значений по хэш-ключу
*/
func ReadCache(hashKey string) (string, error) {
	response, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		return "", err
	}

	return response, nil
}

/*
функция для записи map в кэш
*/
func SaveMapCache(hashKey string, dataMap any) error {
	marshaledMap, err := json.Marshal(dataMap)
	if err != nil {
		return err
	}
	Rdb.Set(Ctx, hashKey, marshaledMap, time.Minute*CacheEXTime).Err()

	return nil
}

/*
Функция для чтения map по хэш-ключу
*/
func ReadMapCache(hashKey string, response any) error {
	cacheData, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}

		return err
	}

	err = json.Unmarshal([]byte(cacheData), &response)
	if err != nil {
		return err
	}

	return nil
}

/*
Функция для удаления значений по хэш-ключу
*/
func DeleteCache(hashKey string) error {
	// Удаляем хэш целиком
	err := Rdb.Del(Ctx, hashKey).Err()
	if err != nil {
		return err
	}
	return nil
}

/*
Функция для удаления значений по шаблону

пример pattern: news_category_*, где * - любое подстановочное значение
*/
func DeleteCacheByPattern(pattern string) error {
	var cursor uint64
	for {
		// Ищем ключи по шаблону
		keys, nextCursor, err := Rdb.Scan(Ctx, cursor, pattern, 10).Result()
		if err != nil {
			return err
		}

		// Удаляем найденные ключи
		if len(keys) > 0 {
			err = Rdb.Del(Ctx, keys...).Err()
			if err != nil {
				return err
			}
		}

		// Обновляем курсор
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

/*
Функция, которая удаляет все протухшие ключ-значения из выбранной таблицы

автоматически применяется при сохранении кэша при помощи функции SaveCache
*/
func DeleteEX(hashKey string) error {
	keys, err := Rdb.HKeys(Ctx, hashKey).Result()
	if err != nil {
		return err
	}

	// удаляем все протухшие ключи из Redis
	for _, key := range keys {
		// Получаем время до истечения срока действия ключа
		ttl := Rdb.TTL(Ctx, key).Val()

		if ttl <= 0 {
			// Если TTL < 0, значит ключ уже истек и можно его удалить
			err := Rdb.Del(Ctx, key).Err()
			if err != nil {
				return err
			}
		}
	}

	return nil
}

/*
функция для стирания кэша

нужна в основном для дэбага
*/
func ClearCache(hashKey string) error {
	// Удаление всего кэша из Redis
	err := Rdb.Del(Ctx, hashKey).Err()
	if err != nil {
		return err
	}
	return nil
}

/*
Функция для чтения значений по хэш-ключу

возвращает grpc response
*/
func ReadProtoCache(hashKey string, m protoreflect.ProtoMessage) (protoreflect.ProtoMessage, error) {
	cacheData, err := Rdb.Get(Ctx, hashKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}

		return nil, err
	}

	err = proto.Unmarshal([]byte(cacheData), m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

/*
функция для записи данных в кэш
*/
func SaveProtoCache(hashKey string, data protoreflect.ProtoMessage) error {
	cacheData, err := proto.Marshal(data)
	if err != nil {
		return nil
	}

	err = Rdb.Set(Ctx, hashKey, cacheData, time.Minute*CacheEXTime).Err()
	if err != nil {
		return err
	}

	return nil
}
