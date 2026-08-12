@echo off
docker compose -f deployments\docker-compose.test.yaml run --rm tests
set TEST_RESULT=%ERRORLEVEL%
docker compose -f deployments\docker-compose.test.yaml down -v
exit /b %TEST_RESULT%