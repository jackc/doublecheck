# Doublecheck

Tool to check data with views that detect invalid data

ideas

- command line tool that checks all tables and prints a report
- rubygem that checks all tables -- designed to integrate with rspec and minitest
- go library that integrates with testing pkg
- tool to write triggers that immediately check just the changed rows
-- might need more metadata to determine what rows are changed and write appropriate triggers
-- maybe store pending checks and have a statement (as opposed to row) trigger on insert to that pending checks table that is defered to transaction commit.
