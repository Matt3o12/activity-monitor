BEGIN;

DROP TABLE IF EXISTS monitors CASCADE;
DROP TABLE IF EXISTS monitor_logs CASCADE;

CREATE TABLE monitors (
    id serial PRIMARY KEY,
    name text NOT NULL,
    type text NOT NULL
);

CREATE TABLE monitor_logs (
    id serial PRIMARY KEY,
    date timestamp with time zone NOT NULL,
    event smallint NOT NULL,
	monitor_id integer  REFERENCES monitors(id) ON DELETE CASCADE
);

INSERT INTO monitors (name, type) VALUES 
	('TCP/UDP Socket', 'socket'),
	('HTTP(s) Server', 'http'),
	('Main Server', 'ping'),
	('Down server', 'ping');

INSERT INTO monitor_logs (date, event, monitor_id) VALUES
	('2016-05-21 21:23:12 Europe/Berlin', 0, 1),
	('2016-05-21 21:23:36 Europe/Berlin', 5, 1),
	('2016-05-22 02:32:10 Europe/Berlin', 5, 2),
	('2016-05-22 02:32:12 Europe/Berlin', 5, 3),
	('2016-05-22 02:32:18 Europe/Berlin', 5, 4),
	('2016-05-22 03:13:20 Europe/Berlin', 4, 1),
	('2016-05-22 03:16:29 Europe/Berlin', 5, 1);

COMMIT;
