# Doublecheck

Doublecheck is both a pattern and a set of tools help preserve data integrity
when domain rules cannot be expressed as foreign key or check constraints.

For example, consider the following domain:

```sql
create table teams(
  id serial primary key,
  name varchar not null,
  max_players integer not null
);

create table players(
  id serial primary key,
  team_id integer not null references teams,
  name varchar not null
);
```

There is a domain rule that the count of players belonging to a team cannot be
greater than `max_players`. The traditional way of enforcing this at the
database layer would be stored procedures to regulate access to the `players`
and `teams` tables or constraint triggers on `teams` and `players` that would
check if the change is allowed. Both of these techniques are imperative rather
than declarative validation. They require knowledge of how to write in a
database procedural language. They also can be error-prone.

An alternative is to use a view that detects invalid data. The maximum team size
constraint can easily be expressed in a view.

```sql
create view teams_with_too_many_players as
select teams.id, teams.name, teams.max_players, count(*)
from teams join players on teams.id=players.team_id
group by 1
having count(*) > teams.max_players;
```

If there are ever any rows in this table, something has gone wrong.

Once this view is created, the issue becomes when do we check the view.

The simplest approach is simply to schedule a check on a regular basis (e.g.
every night). This will not prevent errors from entering the database, but will
detect them quickly. The data can then be corrected and the application bug that
created the invalid data can be fixed. The `doublecheck` command line tool makes
this easy.

The next place that using doublecheck views can be helpful is after every
test case that integrates with the database. This can detect application errors
before they get to production. Also, it is common for the beginning of a test to
setup the database with initial data needed for the test. In complicated
domains, it can be easy to accidently setup the database in an invalid state
which can render the test unreliable. Examining doublecheck views before every
test can detect bad test setup.

For the Ruby language, the
[doublecheck_view](https://github.com/jackc/doublecheck_view) gem encapsulates
this pattern. For other languages, it is simply a matter of hooking into the
test runner's after test functionality and checking that all doublecheck views
have 0 rows.

The previous two approaches have the advantage of not requiring any advanced
database knowledge or code. The disadvantage is they do not absolutely prevent
invalid data from entering the system. For the highest level of data integrity,
it is possible to use doublecheck views with constraint triggers. In this case,
on write to a table that was validated by a doublecheck view the trigger would
ensure the view was empty or it would raise an exception.

## Usage

Install the doublecheck command line tool with `go get`.

```
go get github.com/jackc/doublecheck/cmd/doublecheck
```

Using your preferred migration system or by hand in psql create the schema
`doublecheck`.

```sql
create schema doublecheck;
```

Next create views that detect invalid data in the `doublecheck` schema.

Now to check your database just run `doublecheck check` and a report will be generated in JSON:

```
jack@edi:~$ doublecheck check --database doublecheck_minitest_test
{
  "database": "doublecheck_minitest_test",
  "schema": "doublecheck",
  "user": "jack",
  "start_time": "2016-05-14T10:56:34.629934938-05:00",
  "duration": 1149750,
  "error_detected": true,
  "view_results": [
    {
      "name": "teams_with_too_many_players",
      "start_time": "2016-05-14T10:56:34.629936088-05:00",
      "duration": 1147873,
      "rows": [
        {
          "count": 4,
          "id": 9,
          "max_players": 3,
          "name": "Bears"
        }
      ]
    }
  ]
}
```

Or use the `text` format for easier human consumption:

```
jack@edi:~$ doublecheck check --database doublecheck_minitest_test --format text
Database:   doublecheck_minitest_test
Schema:     doublecheck
User:       jack
Start Time: 2016-05-14 10:57:55.84141567 -0500 CDT
Duration:   475.233µs

---
Name:       teams_with_too_many_players
Start Time: 2016-05-14 10:57:55.841416231 -0500 CDT
Duration:   474.276µs
Error Rows:
  | count: 4 | id: 9 | max_players: 3 | name: Bears |
```

The `-quiet` option can be used to entirely silence output when there are no errors.

Connection options can be specified with command line arguments or via the
standard PG* environment variables.

## Testing

```
createdb doublecheck_test
psql -f setup.sql doublecheck_test
go test
```

If you need to customize the database connection, you can use the standard PG* environment variables.
