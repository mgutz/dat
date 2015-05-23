--@sproc
CREATE FUNCTION add() RETURNS text AS $$
BEGIN
	return 'Hello world!';
END;
$$ LANGUAGE plpgsql;

--@sproc
CREATE FUNCTION subtract() RETURNS text AS $$
BEGIN
	return 'Hello world!';
END;
$$ LANGUAGE plpgsql;

--@key:"foobar"
INSERT INTO user_matches values(s, name)
VALUES ('foo', 'bar');

--@key:"other"
INSERT INTO user_matches values(s, name)
VALUES ('foo', 'bar');

