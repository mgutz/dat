create function f_deposit(_id int, _amount numeric)
returns numeric as $$
declare
	_balance numeric;
begin
	if _amount < 0.01 then
		raise exception 'Amount must be positive';
	end if;

	update accounts
	set balance = balance + _amount
	where id = _id
	returning balance into _balance;

	return _balance;
end;
$$ language plpgsql;

GO

create function f_withdraw(_id int, _amount numeric)
returns numeric as $$
declare
	_balance numeric;
begin
	if _amount < 0.01 then
		raise exception 'Amount must be positive';
	end if;

	select balance into _balance
	from accounts
	where id = _id;

	if _balance - _amount < 0.0 then
		raise exception 'Insufficient balance';
	end if;

	update accounts
	set balance = balance - _amount
	where id = _id
	returning balance into _balance;

	return _balance;
end;
$$ language plpgsql;

