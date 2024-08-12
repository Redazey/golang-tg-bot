INSERT INTO workers (tg_id, name, is_admin, status)
VALUES 
    (887126386, 'Reda', true, false),
    (635983540, 'Tvou Drug/s 💙💛', true, false);

INSERT INTO usercategories (short_name, name, description, data_format, price)
VALUES 
    ('CR', 'Experian', 'simple desc', 'Full name;address;city;state;ZIP;DOB;SSN', 8), 
    ('TU', 'Trans Union', 'simple desc', 'Full name;address;city;state;ZIP;DOB;SSN', 11), 
    ('BG', 'Background', 'simple desc', 'Full name;city;state;dob;age', 8),
    ('fullz', 'Ready Fulls', `Fullz with ready experian in format
name;address;city;state;zip;dob;dl;dl issue date;expiration date
credit score 700+`, '', 8)


