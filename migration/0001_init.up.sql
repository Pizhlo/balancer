CREATE TABLE IF NOT EXISTS config (
	id serial,
	address text,
	rr_weight int,
	is_active boolean
);

INSERT INTO config VALUES(1, 'target1:8081', 1, 'f');
INSERT INTO config VALUES(2, 'target2:8081', 2, 'f');
INSERT INTO config VALUES(3, 'target3:8081', 2, 'f');
INSERT INTO config VALUES(4, 'target4:8081', 1, 'f');
INSERT INTO config VALUES(5, 'target5:8081', 3, 'f');