CREATE TABLE invoices (
    id text primary key not null, -- uuid v7
    client_id text not null,
    invoice_number varchar(50) unique not null,
    period_type varchar(20) not null, -- 'day', 'week', 'fortnight', 'month'
    period_start_date date not null,
    period_end_date date not null,
    subtotal_amount decimal(10,2) not null default 0.00,
    gst_amount decimal(10,2) not null default 0.00,
    total_amount decimal(10,2) not null default 0.00,
    generated_date datetime default current_timestamp not null,
    created_at datetime default current_timestamp not null,
    updated_at datetime default current_timestamp not null,
    foreign key (client_id) references clients(id)
);

alter table sessions add column invoice_id text;

create index idx_invoices_client_id on invoices(client_id);
create index idx_invoices_invoice_number on invoices(invoice_number);
create index idx_invoices_period_dates on invoices(period_start_date, period_end_date);
create index idx_sessions_invoice_id on sessions(invoice_id);

create trigger invoices_updated_at 
    after update on invoices 
    begin
        update invoices set updated_at = current_timestamp where id = new.id;
    end;

create table payments (
	id text primary key not null, -- uuid v7
	invoice_id text not null,
	amount decimal(10,2) not null,
	payment_date date not null,
	created_at datetime default current_timestamp not null,
	updated_at datetime default current_timestamp not null,
	foreign key (invoice_id) references invoices(id)
);

create view v_invoices as
select 
	i.*,
	cast(coalesce(sum(p.amount), 0.0) as real) as amount_paid,
	max(p.payment_date) as payment_date
from invoices i
left join payments p on p.invoice_id = i.id
group by i.id;
