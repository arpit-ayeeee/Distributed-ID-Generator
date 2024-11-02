package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	numBuckets = 2
)
const (
	dbUser     = ""
	dbPassword = ""
	dbName     = ""
)

// Read the file on reboot, add the modulus and safety buffer
var goroutineCounter int64 = readCounterValue() + 100 + 5

func readCounterValue() int64 {
	// Read the content of the file
	// Open the file
	file, err := os.Open("counter.txt")
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return 0
	}
	defer file.Close()

	// Read the file content
	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Println("Error getting file info:", err)
		return 0
	}

	data := make([]byte, fileInfo.Size())
	_, err = file.Read(data)
	if err != nil {
		fmt.Println("Error reading the file:", err)
		return 0
	}

	counterStr := string(data)
	counterValue, err := strconv.ParseInt(counterStr, 10, 64)
	if err != nil {
		fmt.Println("Error converting to int64:", err)
		return 0
	}

	fmt.Println("Fetched integer value from file:", counterValue)
	return counterValue
}

func writeIntoFile(counterValue int64) {

	file, err := os.OpenFile("counter.txt", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("Error creating or opening the file:", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(strconv.FormatInt(counterValue, 10))
	if err != nil {
		fmt.Println("Error writing to the file:", err)
		return
	}
	fmt.Println("Content written to file successfully.")
}

// Basic implementation which handles uniqueness
//   - multiple machineId, basis which servers sends the request
//   - counter which atomically increments, i.e. handling of thread
//   - persisting counter value in file, in case of reboot
func getNewDistributedId(threadId int64, machineId string) string {
	epochTime := time.Now().UnixNano() / int64(time.Millisecond)

	// Combine the epoch time, machine ID, and thread ID into a single string
	uniqueID := fmt.Sprintf("%d-%s-%d", epochTime, machineId, threadId)

	if threadId%100 == 0 {
		fmt.Printf("Writing in file: %d", threadId)
		writeIntoFile(threadId)
	}

	return uniqueID
}

func getShardIndex(userId string) int {
	userIdStr := userId
	hasher := sha256.New()
	hasher.Write([]byte(userIdStr))
	hashBytes := hasher.Sum(nil)
	hashInt := new(big.Int).SetBytes(hashBytes)
	bucket := new(big.Int).Mod(hashInt, big.NewInt(int64(numBuckets)))
	return int(bucket.Int64())
}

func addUserDetails(wg *sync.WaitGroup, message string) {
	defer wg.Done()

	//Implement file persisted atomic counter, so that it's value doesn't sets to 0 on reboot
	goroutineID := atomic.AddInt64(&goroutineCounter, 1)

	userId := getNewDistributedId(goroutineID, "m1")
	shardIndex := getShardIndex(userId)

	dsn := fmt.Sprintf("%s:%s@tcp(127.0.0.1:3306)/%s", dbUser, dbPassword, dbName+strconv.Itoa(shardIndex+1))

	fmt.Println(dsn, goroutineID, shardIndex)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Error opening DB: %v\n", err)
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO Messages (id, message) VALUES (?, ?)", userId, message)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	var wg sync.WaitGroup

	userIdMap := make(map[int][]int)
	fmt.Println("Adding Users to DB...")
	for i := 1; i <= 1000; i++ {
		wg.Add(1)
		go addUserDetails(&wg, "message"+strconv.Itoa(i))
	}
	wg.Wait()

	fmt.Println("Adding Users completed: ", userIdMap)

	wg.Wait()
}
