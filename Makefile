export PGUSER=postgres
export PGPASSWORD=postgres

test: db.reset
	@go test -count=1 ./...

test.cover: db.reset
	@mkdir -p .coverage
	@go test -coverprofile .coverage/report.out . > /dev/null
	@go tool cover -func=.coverage/report.out -o=.coverage/report.text
	@go tool cover -html=.coverage/report.out -o=.coverage/report.html
	@cat .coverage/report.text

db.reset:
	@psql -tc "select 'drop database pmx' from pg_database where datname = 'pmx'" \
		|xargs -I{statement} psql -tc "{statement}" > /dev/null
	@psql -c 'create database pmx' > /dev/null
	@psql -d pmx -f fxt/schema.sql > /dev/null
