PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;

CREATE TABLE IF NOT EXISTS majorMajorCat (
    code TEXT PRIMARY KEY,
    name TEXT,
    blurb TEXT,
    sort INT
);

CREATE TABLE IF NOT EXISTS majorCat (
    majorMajorCatCode TEXT,
    code TEXT,
    name TEXT,
    sort INT,
    PRIMARY KEY (majorMajorCatCode, code)
);

CREATE TABLE IF NOT EXISTS MinorCat (
    majorMajorCatCode TEXT,
    majorCatCode TEXT,
    code TEXT,
    name TEXT,
    sort INT,
    PRIMARY KEY (majorMajorCatCode, majorCatCode, code)
);

REPLACE INTO majorMajorCat(code, name, blurb, sort) VALUES
    ('services', 'Services', 'Skills, talents and services useful to advise and shape digital futures.', 0),
    ('products', 'Products', 'Hardware, software and accessories to make digital happen.', 1);

REPLACE INTO majorCat(majorMajorCatCode, code, name, sort) VALUES
('services','web','Web Development',1),
('services','hosting','Website Hosting',8),
('services','online-strategy','Online Strategy / Planning',9),
('services','social','Social Media / Communities Support',10),
('services','promo','Online Promotion / Advertising',11),
('services','seo','Search Engine Optimisation',12),
('services','mobile','Mobile / Apps Development',13),
('services','dev','Software Development / Programming',14),
('services','ba','Business Analysis',15),
('services','pm','Project Management',16),
('services','analysis','Data Analysis / Business Intelligence',17),
('services','doc','Technical Documentation Writing / Editing',18),
('services','dba','Database Administration',19),
('services','network','Networking & Information Systems',20),
('services','security','Network / Device Security',21),
('services','backup','Data Backup',22),
('services','analysis','Data Analysis / Business Intelligence',23),
('services','telco','Phone Systems Implementation',24),
('services','cabwire','Cabling / Wireless',27),
('services','comprepair','Computer Repairs / Support',28),
('services','electron','Electronics Repairs / Servicing',29),
('services','edu','Education & Training ',30),
('services','design','Graphic Design',32),
('services','digprint','Digital Printing',33),
('services','anim','Animation',34),
('services','interact','Interactive Design',35),
('services','photo','Photos',36),
('services','video','Digital Video',40),
('services','fab','Digital Fabrication / Making',41),
('services','coach','Digital Futures / Innovation Coaching',45),
('services','governance','Governance',46),
('services','ip','Intellectual Property / Rights Advice',47);

REPLACE INTO minorCat(majorMajorCatCode, majorCatCode, code, name, sort) VALUES
('services','web','site','Site Development',2),
('services','web','arch','Information Architecture / Usability',3),
('services','web','ecom','eCommerce / Shopping Carts',4),
('services','web','cms','Content Management Systems',5),
('services','web','copy','Web Content Writing / Editing',6),
('services','web','optimisation','Web Optimisation / Reporting',7),
('services','telco','telcoplan','System Planning',25),
('services','telco','telcosupport','System Support',26),
('services','edu','courseware','Courseware Design',31),
('services','photo','printing','Photography',37),
('services','photo','3dtours','3D Tours',38),
('services','photo','3dtours','Digital Photo Printing',39),
('services','fab','3dmodel','3D Modelling',42),
('services','fab','3dscan','3D Scanning',43),
('services','fab','3dprinting','3D Printing',44);



CREATE TABLE IF NOT EXISTS listing (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    Status INT DEFAULT 0,
    AdminEmail TEXT,
    AdminFirstName TEXT,
    AdminLastName TEXT,
    AdminPhone TEXT,
    WCCExportOK INT,
    isOrg INT,
    Name TEXT,
    Desc1 TEXT,
    Desc2 TEXT,
    Phone TEXT,
    Email TEXT,
    Websites TEXT,
    Address TEXT,
    ImageId TEXT,
    Updated DATETIME
);

CREATE INDEX IF NOT EXISTS listingStatus ON listing(Status);

CREATE INDEX IF NOT EXISTS listingIsOrg ON listing(isOrg);

CREATE TABLE IF NOT EXISTS categoryListing (
    majorMajorCatCode TEXT,
    majorCatCode TEXT,
    minorCatCode TEXT,
    listingId INT,
    PRIMARY KEY (majorMajorCatCode, majorCatCode, minorCatCode, listingId)
);

CREATE TRIGGER IF NOT EXISTS deleteListingsCategory AFTER DELETE ON listing
  BEGIN
    DELETE FROM categoryListing WHERE listingId = old.id;
  END;

CREATE TABLE IF NOT EXISTS session (
    id TEXT PRIMARY KEY,
    data BLOB,
    expires TEXT
);


CREATE TABLE IF NOT EXISTS login (
    code TEXT PRIMARY KEY,
    email TEXT,
    expires DATETIME
);

CREATE VIRTUAL TABLE IF NOT EXISTS listing_fts  USING fts4(content="listing", tokenize=porter, Name, Desc1, Desc2);

CREATE TRIGGER IF NOT EXISTS listing_bu BEFORE UPDATE ON listing BEGIN
  DELETE FROM listing_fts WHERE docid=old.rowid;
END;
CREATE TRIGGER IF NOT EXISTS listing_bd BEFORE DELETE ON listing BEGIN
  DELETE FROM listing_fts WHERE docid=old.rowid;
END;

CREATE TRIGGER IF NOT EXISTS listing_au AFTER UPDATE ON listing BEGIN
  INSERT INTO listing_fts(docid, Name, Desc1, Desc2) VALUES(new.rowid, new.Name, new.Desc1, new.Desc2);
END;
CREATE TRIGGER IF NOT EXISTS listing_ai AFTER INSERT ON listing BEGIN
  INSERT INTO listing_fts(docid, Name, Desc1, Desc2) VALUES(new.rowid, new.Name, new.Desc1, new.Desc2);
END;

INSERT INTO listing_fts(listing_fts) VALUES('rebuild');

CREATE TABLE IF NOT EXISTS image (
    id TEXT PRIMARY KEY,
    created DATETIME,
    format TEXT,
    original BLOB,
    small BLOB,
    large BLOB
);