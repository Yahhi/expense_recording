# Gastos_ingresos_bot
Go-lang project for saving personal expenses and be up-to-date with budgeting

## Database
The database used is Postgres.

Database upgrades can be done with a command:
```
$ migrate -source db/migrations -database postgres://<database connection> up 2
```
All current table creation scripts are located in db/migrations

More information: https://github.com/golang-migrate/migrate

## Health check
There is a default endpoint for healthcheck which replies OK when asked as a web-server at port 8080

## Running
You have to set up the following settings as environment variables:
```
PGHOST=<HOST> 
PGPORT=<PORT>
PGADMIN=<USER>
PGPASS=<PASSWORD>
PGDBNAME=<DBNAME>
TGTOKEN=<TELEGRAM-BOT-TOKEN>
```
So, everything you need to run it:
- database connection settings
- telegram API token for bot (can be obtained from Bot Father when you create a bot)

## Additional info
Miro board with the schema of functions:
https://miro.com/app/board/uXjVNjQKds8=/
