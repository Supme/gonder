ALTER TABLE sender
    ADD utm_url VARCHAR(100) DEFAULT '' NOT NULL AFTER name;