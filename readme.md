# Distributed ID Generation with Persistent Counter

This Go package generates unique distributed IDs by combining the current timestamp, machine ID, and atomic counter values. It also provides functionality for sharding based on generated IDs and storing user data across a distributed system of database shards.

## Features

- **Atomic Counter**: Implements an atomic counter that persists across application restarts by reading/writing to a file.
- **Sharding**: Provides a sharding mechanism based on a user ID hash function.
- **Unique ID Generation**: Combines epoch timestamp, machine ID, and counter values to generate globally unique IDs.
- **File-based Counter Persistence**: Reads and writes counter values to `counter.txt`, ensuring counter continuity after reboots.
- **Database Sharding**: Distributes user data across multiple database shards using consistent hashing.

## Files

- `counter.txt`: Stores the atomic counter value, enabling persistence across reboots.
- `main.go`: Main application file that includes ID generation, sharding, and database insertion logic.

## Prerequisites

1. Go 1.16+
2. MySQL database (Install MySQL driver for Go)
3. Set up environment variables in the code for `dbUser`, `dbPassword`, and `dbName`.

## Installation

To get started, clone the repository and install the dependencies:

```sh
git clone <repository-url>
cd <repository-directory>
go mod init
go mod tidy
```

## Configuration

Edit the database connection constants in `main.go`:

```go
const (
    dbUser     = "<YOUR_DB_USER>"
    dbPassword = "<YOUR_DB_PASSWORD>"
    dbName     = "<YOUR_DB_NAME>"
)
```

Ensure that MySQL is accessible on `127.0.0.1:3306` and create the database with a table `Messages`:

```sql
CREATE TABLE Messages (
    id VARCHAR(255) PRIMARY KEY,
    message TEXT
);
```

## Usage

The application reads the `counter.txt` file on start-up to initialize the atomic counter. If `counter.txt` is missing, it initializes the counter to zero.

### Running the Application

To run the application, use the following command:

```sh
go run main.go
```

### Functions

1. **`getNewDistributedId`**: Combines epoch time, machine ID, and thread ID to create a unique ID.
2. **`getShardIndex`**: Uses SHA-256 hashing to calculate the database shard based on the user ID.
3. **`addUserDetails`**: Inserts user data into a database shard based on the calculated shard index.

### Example Output

```plaintext
Adding Users to DB...
Starting Insert on DB Shard 1 for User ID: 20230101-m1-1
Starting Insert on DB Shard 2 for User ID: 20230101-m1-2
...
Adding Users completed.
```

## File-based Counter Persistence

The application reads the counter from `counter.txt` and writes updates to this file when a counter crosses specific thresholds.

- **File Reading**: `readCounterValue` reads the persisted counter from `counter.txt`.
- **File Writing**: `writeIntoFile` updates `counter.txt` with the latest counter value when specific conditions are met.
