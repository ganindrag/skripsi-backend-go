create table programmer(id bigserial primary key not null, name varchar(100) not null, email varchar(100) not null, password varchar(255) not null, role varchar(5) not null);

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

create table feedback(
	id bigserial primary key not null,
	programmer_id bigint,
	task_id bigint,
	feedback text,
	created_at timestamp default now(),
	CONSTRAINT fk_programmer_feedback
	FOREIGN KEY(programmer_id) 
	REFERENCES programmer(id),
	CONSTRAINT fk_tugas_feedback
	FOREIGN KEY(task_id) 
	REFERENCES task(id)
);

insert into programmer
	values (default, 'admin', 'admin', md5('admin'), 'ADMIN'),
(default, 'gege', 'gege@gmail.com', md5('gege'), 'MNGR'),
(default, 'joko', 'joko@gmail.com', md5('joko'), 'USER'),
(default, 'rizki', 'rizki@gmail.com', md5('rizki'), 'USER');

