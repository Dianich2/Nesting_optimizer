package postgres

const existUserByLoginQuery = `
	select exists(
		select 1 
		from users 
		where login = $1 
			and deleted_at is null
	);
`

const existUserByEmailQuery = `
	select exists(
		select 1 
		from users 
		where email = $1 
		    and deleted_at is null
	);
`

const createUserQuery = `
	insert into users(login, email, password_hash, first_name, last_name)
	values($1, $2, $3, $4, $5)
	returning id, login, email, first_name, last_name, created_at, updated_at;
`
