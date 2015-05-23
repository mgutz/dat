--@sproc=Add
CREATE FUNCTION add() RETURNS text AS $$
BEGIN
	return 'Hello world!';
END;
$$ LANGUAGE plpgsql;

--@sproc=Subtract
CREATE FUNCTION subtract() RETURNS text AS $$
BEGIN
	return 'Hello world!';
END;
$$ LANGUAGE plpgsql;

--@key=BuildFoobar
INSERT INTO user_matches values(s, name)
VALUES ('foo', 'bar');

--@key=InsertUsers
INSERT INTO users
VALUES ($1, $2);

--@key=foobar

