INSERT INTO workers (tg_id, is_admin, status)
VALUES 
    (1, true, false),
    (67890, false, true),
    (11111, true, false),
    (22222, false, true),
    (33333, true, false);

INSERT INTO tickets (tg_id, status, "timestamp")
VALUES 
    (12345, 'in_progress', '2022-01-01 12:00:00'),
    (67890, 'done', '2022-01-02 13:00:00'),
    (11111, 'in_progress', '2022-01-03 14:00:00'),
    (22222, 'done', '2022-01-04 15:00:00'),
    (33333, 'in_progress', '2022-01-05 16:00:00');

INSERT INTO usercategories (name)
VALUES ('CR'), ('TU');


