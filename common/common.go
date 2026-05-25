package common

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/joho/godotenv"
	"github.com/kurmidev/annualstatement/data"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
)

type Common struct {
	DB     *db.Session
	S3Los  *session.Session
	S3Sftp *session.Session
	FinYrs string
}

func New(finStartDate time.Time) Common {
	c := Common{}
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Unable to load the data from env file", err)
	}
	//Assigning the db connections
	settings := mysql.ConnectionURL{
		Database: os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
	}

	db, err := c.dbConnect(settings)
	if err != nil {
		log.Fatal("Erorr connecting Database", err)
	}

	c.DB = &db

	//connecting the S3 for upload and download of PDF files
	s3Los, err := c.S3Connections(os.Getenv("AWS_REGION"), os.Getenv("AWS_LOS_ID"), os.Getenv("AWS_LOS_SECRET"))
	if err != nil {
		log.Fatal("Erorr connecting LOS S3", err)
	}

	c.S3Los = s3Los

	s3Sftp, err := c.S3Connections(os.Getenv("AWS_REGION"), os.Getenv("AWS_SFTP_ID"), os.Getenv("AWS_SFTP_SECRET"))
	if err != nil {
		log.Fatal("Erorr connecting SFTP S3", err)
	}
	c.S3Sftp = s3Sftp
	c.FinYrs = c.CalculateFinyrs(finStartDate)
	return c
}

func (c Common) dbConnect(set mysql.ConnectionURL) (db.Session, error) {
	sess, err := mysql.Open(set)
	if err != nil {
		log.Fatal("Error connecting database", err)
	}
	return sess, err
}

func (c Common) S3Connections(region, id, secret string) (*session.Session, error) {
	return session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(id, secret, ""),
	})
}

func (c Common) DownloadFile(fileName string) (*os.File, error) {
	fmt.Println("Searching file", fileName)
	file, err := os.Create(fileName)
	if err != nil {
		log.Fatalf("Unable to open file %q, %v", fileName, err)
		return nil, err
	}

	downloader := s3manager.NewDownloader(c.S3Los)
	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(os.Getenv("LOS_BUCKET")),
			Key:    aws.String(fileName),
		})

	if err != nil {
		log.Fatalf("Unable to download item %q, %v", fileName, err)
		file.Close()
		return nil, err
	}
	return file, nil
}

func (c Common) UploadFile(file *os.File, s3Folder, fileName string) (bool, error) {
	fileDir := fmt.Sprintf("/%s/Annual-statement/FY-%s/%s", s3Folder, c.FinYrs, fileName)
	fmt.Printf("Uploading file %s to new path %s \n", fileName, fileDir)
	fileInfo, _ := file.Stat()
	fileSize := fileInfo.Size()
	fileBuffer := make([]byte, fileSize)
	file.Read(fileBuffer)
	_, err := s3.New(c.S3Sftp).PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(os.Getenv("SFTP_BUCKET")),
		Key:           aws.String(fileDir),
		Body:          bytes.NewReader(fileBuffer),
		ContentLength: aws.Int64(fileSize),
		ContentType:   aws.String(http.DetectContentType(fileBuffer)),
	})

	defer func(file *os.File, fileName string) {
		file.Close()
		os.Remove(fileName)
	}(file, fileName)

	if err != nil {
		log.Fatal("Uploading file failed \n", err)
		return false, err
	}
	fmt.Printf("File uploaded successfully %s \n", fileName)
	return true, nil
}

func (c Common) CalculateFinyrs(finYrs time.Time) string {
	fin := ""
	currentYrs := finYrs.Year() % 1e2
	if finYrs.Month() >= 4 {
		fin = fmt.Sprintf("%d-%d", currentYrs, (currentYrs + 1))
	} else {
		fin = fmt.Sprintf("%d-%d", (currentYrs - 1), currentYrs)
	}
	return fin
}

func (c Common) PerformSyncOld(statements []data.InvInvestorStatementLog, s3Folder string) (int, int) {
	successCnt := 0
	errorCnt := 0
	for _, statement := range statements {
		file, err := c.DownloadFile(statement.FileName)
		if err != nil {
			errorCnt++
		}
		resp, err := c.UploadFile(file, s3Folder, statement.FileName)
		if resp {
			successCnt++
		}
		if err != nil {
			errorCnt++
		}
	}

	return successCnt, errorCnt
}

func (c Common) PerformSync(statements []data.InvInvestorStatementLog, s3Folder string) (int, int) {
	const maxConcurrent = 10 // Limit concurrency to prevent "too many open files"
	var (
		successCnt int
		errorCnt   int
		wg         sync.WaitGroup
		mu         sync.Mutex                           // to protect counters
		sem        = make(chan struct{}, maxConcurrent) // semaphore
	)

	for _, statement := range statements {
		wg.Add(1)
		sem <- struct{}{} // acquire a slot

		go func(stmt data.InvInvestorStatementLog) {
			defer wg.Done()
			defer func() { <-sem }() // release the slot when done

			file, err := c.DownloadFile(stmt.FileName)
			if err != nil {
				mu.Lock()
				errorCnt++
				mu.Unlock()
				return
			}

			resp, err := c.UploadFile(file, s3Folder, stmt.FileName)
			mu.Lock()
			if err != nil {
				errorCnt++
			} else if resp {
				successCnt++
			}
			mu.Unlock()
		}(statement)
	}

	wg.Wait() // wait for all goroutines to finish

	return successCnt, errorCnt
}
