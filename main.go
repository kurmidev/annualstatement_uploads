package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/kurmidev/annualstatement/common"
	"github.com/kurmidev/annualstatement/data"
)

var Client = map[int]string{
	1799: "Prudent",
	3569: "ET-Money-Supply",
	1827: "Epifi",
	2300: "Uni",
	436:  "Fincart",
	3597: "Centricity",
	1596: "Abchlor",
}

func main() {

	ifaId, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("invalid input parameter ", err)
	}

	finStartDate, err := time.Parse("2006-01-02", os.Args[2])
	if err != nil {
		log.Fatal("Invalid start date of the finnacial yearsit should be YYYY-mm-dd format ", err)
	}

	fmt.Println(ifaId, finStartDate)
	c := common.New(finStartDate)
	m := data.New(ifaId, finStartDate, *c.DB)
	var statements []data.InvInvestorStatementLog
	statements, err = m.GetStatementData()
	fmt.Println("Total investor count ", len(statements))

	if len(statements) <= 0 {
		log.Fatal("Not able to find the Child IFA ", err)
	} else {
		s3Folder := Client[ifaId]
		success, errorc := c.PerformSync(statements, s3Folder)
		fmt.Printf("current upload status succes %d and error is %d", success, errorc)
	}

}
