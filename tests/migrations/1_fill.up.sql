-- Вставка 5 тестовых пользователей
INSERT INTO users (tg_id, limits)
VALUES (1001, 10000), -- limits в копейках (100 рублей)
       (1002, 20000),
       (1003, 30000),
       (1004, 40000),
       (1005, 50000);

-- Создание тестовых категорий (предполагаем, что таблица usercategories уже существует)
-- Если нет, создайте ее перед запуском скрипта.
INSERT INTO usercategories (name) 
VALUES ('TU'), 
       ('CR'), 
       ('FULLZ'), 

-- Вставка 5 тестовых транзакций для каждого пользователя
INSERT INTO usermoneytransactions (user_id, category_id, sum, "timestamp")
SELECT 
    u.id, -- user_id
    floor(random() * 3) + 1, -- случайная категория от 1 до 3
    floor(random() * 1000), -- случайная сумма от 0 до 1000
    now() - (floor(random() * 30) || ' days')::interval -- случайная дата в пределах последних 30 дней
FROM users u
CROSS JOIN LATERAL generate_series(1, 5);