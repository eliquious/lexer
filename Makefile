default: test

test:
	@echo "------------------"
	@echo " test"
	@echo "------------------"
	@go test -coverprofile=coverage.out

benchmark:
	@echo "------------------"
	@echo " benchmark"
	@echo "------------------"
	@go test -test.bench="^Bench*"

coverage: test
	@echo "------------------"
	@echo " coverage report"
	@echo "------------------"
	@go tool cover -html=coverage.out -o coverage.html

html:
	@echo "------------------"
	@echo " html report"
	@echo "------------------"
	@go tool cover -html=coverage.out -o coverage.html
	@open coverage.html

detail:
	@echo "------------------"
	@echo " detailed report"
	@echo "------------------"
	@gocov test | gocov report

report: test detail html

