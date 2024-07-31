-- Drop indexes first
DROP INDEX IF EXISTS workers_tg_id;
DROP INDEX IF EXISTS usermoneytransactions_user_id_period;
DROP INDEX IF EXISTS usercategories_lower_name;
DROP INDEX IF EXISTS users_tg_id;
DROP INDEX IF EXISTS tickets_id;

-- Drop tables
DROP TABLE IF EXISTS usermoneytransactions;
DROP TABLE IF EXISTS usercategories;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS workers;
DROP TABLE IF EXISTS tickets;