default := 'all'
# Build the 'gen' binary
build-gen:
    echo "Building gen binary..."
    go build -o gen dataset-gen/main.go

# Build the 'load_test' binary
build-load-test:
    echo "Building load_test binary..."
    go build -o load_test api-load-testing/main.go

# Generate database command (automatically depends on building gen)
generate-dataset: build-gen
    echo "Generating database with num=$((7150000*8)) and parallel=4"
    ./gen --num=$((7150000*8)) --parallel=4

# Load test command (automatically depends on building load_test)
load-test *ARGS: build-load-test
    echo "Starting Docker containers..."
    docker-compose up -d
    echo "Running load test with args {{ARGS}}"
    ./load_test {{ARGS}}

# Stop Docker containers
stop-docker:
    echo "Stopping Docker containers..."
    docker-compose down --volumes
