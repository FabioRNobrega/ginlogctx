.PHONY: docker-test

docker-test:
	@echo "Running ginlogctx tests in Docker..."
	docker compose run --rm test
	@echo "ginlogctx Docker tests passed."
