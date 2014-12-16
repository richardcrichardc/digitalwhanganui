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
('services','online-strategy','Online Strategy / Planning',6),
('services','copy','Web Content Writing / Editing',7),
('services','optimisation','Web Optimisation / QA / Reporting',8),
('services','social','Social Media / Communities Support',9),
('services','mobile','Mobile / Apps Development',10),
('services','promo','Online Promotion / Advertising / Growth Hacking',11),
('services','dev','Software Development',12),
('services','ba','Business Analysis',14),
('services','pm','Project Management',15),
('services','doc','Technical Documentation Writing / Editing',16),
('services','dba','Database Administration',17),
('services','network','Networking & Information Systems',18),
('services','analysis','Data Analysis / Business Intelligence',19),
('services','telco','Phone Systems Support?',20),
('services','edu','Education & Training ',21),
('services','design','Graphic Design',23),
('services','anim','Animation',24),
('services','photo','Photography',25),
('services','printing','Printing from Digital ',27),
('services','video','Video Capture / Editing',28),
('services','3d','Engineering / Architecture / Making',29),
('services','strategy','Digital Strategy / Planning',33),
('services','coach','Digital Futures / Innovation Coaching',34),
('services','ip','Intellectual Property?',35),
('products','comp','Computers',1),
('products','comp-acc','Computer Accessories',2),
('products','software','Software',3),
('products','tablets','Tablets',4),
('products','phones','Phones',5),
('products','phone-acc','Phone Accessories',6),
('products','camera','Cameras, Digital',7),
('products','video','Video Cameras, Digital',8),
('products','games','Games',9),
('products','console','Gaming systems',10),
('products','drones','Drones / Multicopters',11),
('products','fab','Digital Fabrication Tools',12),
('products','electronics','Electronics Supplies',17),
('products','electronics','',18);

REPLACE INTO minorCat(majorMajorCatCode, majorCatCode, code, name, sort) VALUES
('services','web','site','Site Development',2),
('services','web','arch','Information Architecture / Usability / User Experience',3),
('services','web','ecom','eCommerce / Shopping Cart Setup',4),
('services','web','cms','Content Management System Selection',5),
('services','dev','breakdown','(languages breakdown?)',13),
('services','edu','breakdown','(subjects breakdown?)',22),
('services','photo','printing','Digital Photo Printing',26),
('services','3d','scan','3D Scanning',30),
('services','3d','cad','3D Modelling / Product Design (CAD)',31),
('services','3d','printing','3D Printing / Digital Fabrication',32),
('products','fab','3dprint','3D Printers',13),
('products','fab','cnc','CNC Torch Tables',14),
('products','fab','cutters','Digital Cutters',15),
('products','fab','sew','Digital Sewing Machines / Embroidery',16);

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