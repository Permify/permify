# Race Condition Integration Test

This test validates the fix for concurrent transaction race condition in Permify.

## 🚀 Quick Start

### 1. Start Test Environment

```bash
cd integration-test/race_condition
docker compose -f docker-compose-test.yml up --build -d
```

### 2. Run Tests

```bash
# Simple test
./test_race_condition.sh

# Parallel test (more aggressive)
./test_race_condition_parallel.sh
```

## 🧹 Cleanup

```bash
docker compose -f docker-compose-test.yml down -v
```

## 📁 Files

- `schema.perm` - Test schema
- `test_race_condition.sh` - Simple test
- `test_race_condition_parallel.sh` - Parallel test
- `docker-compose-test.yml` - Test environment