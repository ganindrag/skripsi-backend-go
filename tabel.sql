create table programmer(id bigserial primary key not null, name varchar(100), email varchar(100), password varchar(255), role varchar(5));

create table task(
	id bigserial primary key not null, 
	programmer_id bigint, 
	name varchar(255), 
	detail text, 
	status varchar(25) not null, 
	weight int default 1, 
	created_at timestamp default now(), 
	start_at timestamp, 
	end_at timestamp, 
	bug_tolerance int default 0,
	actual_bug int default 0,
	comprehension int default 100,
	is_evaluated boolean default false,
	CONSTRAINT fk_programmer_tugas
	FOREIGN KEY(programmer_id) 
	REFERENCES programmer(id)
);

insert into programmer
	values (default, 'admin', 'admin', md5('admin'), 'ADMIN'),
(default, 'gege', 'gege@gmail.com', md5('gege'), 'ADMIN'),
(default, 'joko', 'joko@gmail.com', md5('joko'), 'USER'),
(default, 'hhowe', 'hhowe@hotmail.com', md5(random()::text), 'USER');

