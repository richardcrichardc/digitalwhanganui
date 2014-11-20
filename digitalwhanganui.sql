CREATE TABLE IF NOT EXISTS majorCat (
    code TEXT PRIMARY KEY,
    name TEXT,
    sort INT
);

CREATE TABLE IF NOT EXISTS MinorCat (
    majorCatCode TEXT,
    code TEXT,
    name TEXT,
    sort INT,
    PRIMARY KEY (majorCatCode, code)
);

REPLACE INTO majorCat(code, name, sort) VALUES
    ('people', 'People', 0),
    ('org', 'Organisations', 1),
    ('animals', 'Animals', 3),
    ('et', 'Extraterestrials', 4);

REPLACE INTO minorCat(majorCatCode, code, name, sort) VALUES
    ('people', 'programmers', 'Programmers', 0),
    ('people', 'designers', 'Designers', 1),
    ('people', 'copywriters', 'Copy Writers', 2),
    ('org', 'shops', 'Shops', 0),
    ('org', 'charities', 'Charities', 1),
    ('org', 'consultancies', 'Consultancies', 2),
    ('animals', 'cats', 'Cats', 0),
    ('animals', 'dogs', 'Dogs', 1),
    ('animals', 'birds', 'Birds', 2),
    ('et', 'martians', 'Martians', 0),
    ('et', 'venetians', 'Venitians', 1),
    ('et', 'sontarans', 'Sontanrans', 2);


CREATE TABLE IF NOT EXISTS listing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    AdminEmail TEXT UNIQUE,
    AdminFirstName TEXT,
    AdminLastName TEXT,
    AdminPhone TEXT,
    Name TEXT,
    Desc1 TEXT,
    Desc2 TEXT,
    Phone TEXT,
    Email TEXT,
    Websites TEXT,
    Address TEXT
);

CREATE TABLE IF NOT EXISTS categoryListing (
    majorCatCode TEXT,
    minorCatCode TEXT,
    listingId INT,
    PRIMARY KEY (majorCatCode, minorCatCode, listingId)
);

CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY,
    data BLOB,
    expires TEXT
);