drop schema if exists doublecheck cascade;
create schema doublecheck;

set search_path = 'doublecheck';

create view without_errors as
select 42::integer as id, 'no-op'::text as error_message
where false;

create view with_multiple_errors as
select 7::integer as id, 'something went wrong'::text as error_message
union all
select 42::integer as id, 'something else went wrong' as error_message
;

-- Test that bad view names can't be used for SQL injection
create view "syntax error" as
select 42::integer as id, 'no-op'::text as error_message
where false;
