package data

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/upper/db/v4"
)

type Model struct {
	IfaId        int
	FinStartDate time.Time
	db           *db.Session
}

func New(ifaId int, finStartDate time.Time, db db.Session) *Model {
	return &Model{IfaId: ifaId, FinStartDate: finStartDate, db: &db}
}

func (m *Model) GetChildIfas() ([]string, error) {
	var childIfa []string

	sqlQuery := fmt.Sprintf(`SELECT group_concat(DISTINCT m1.user_id) as childifa
	FROM      sys_acl_user_role_mapping m1
	LEFT JOIN sys_acl_user_role_mapping m2 ON m2.user_id = m1.reporting_user_id AND m2.role_id IN(12,13,14,23,24)
	LEFT JOIN sys_acl_user_role_mapping m3 ON m3.user_id = m2.reporting_user_id AND m3.role_id IN(12,13,14,23,24)
	LEFT JOIN sys_acl_user_role_mapping m4 ON m4.user_id = m3.reporting_user_id AND m4.role_id IN(12,13,14,23,24)
	LEFT JOIN sys_acl_user_role_mapping m5 ON m5.user_id = m4.reporting_user_id AND m5.role_id IN(12,13,14,23,24)
	LEFT JOIN sys_acl_user_role_master rm ON m1.role_id=rm.id
	LEFT JOIN ifa_user ifu ON m1.user_id = ifu.id
	LEFT JOIN ifa_user ifu2 ON m1.reporting_user_id = ifu2.id
	LEFT JOIN sys_acl_user_role_master rm2 ON m2.role_id=rm2.id
	WHERE 
	(m1.user_id = %d OR %d IN (m1.reporting_user_id, 
			m2.reporting_user_id, 
			m3.reporting_user_id, 
			m4.reporting_user_id, 
			m5.reporting_user_id) )
	AND m1.role_id IN(12,13,14,23,24)`, m.IfaId, m.IfaId)

	rows, err := (*m.db).SQL().Query(sqlQuery)
	if err != nil {
		log.Fatal("Error in query exections: ", err)
		return childIfa, err
	}
	if !rows.Next() {
		log.Fatal("Expecting atleast  one row in output")
		return childIfa, err
	}

	var response string
	if err := rows.Scan(&response); err != nil {
		log.Fatal("Error getting record from table", err)
		return childIfa, err
	}

	childIfa = strings.Split(response, ",")
	return childIfa, nil
}

func (m *Model) GetStatementData() ([]InvInvestorStatementLog, error) {
	var statements []InvInvestorStatementLog
	childifa, err := m.GetChildIfas()
	if err != nil {
		log.Fatal("Error getting child and master ifa", err)
		return statements, err
	}
	q := (*m.db).SQL().Select("a.*").
		From("inv_investor_statement_log a ").
		Join("inv_investor_user u ").On("a.investor_id=u.id").
		Where("a.statement_type =?", "AnnualStatement").
		Where("a.start_date =?", m.FinStartDate).
		Where("a.status=?", "Active").
		Where("a.is_deleted=?", "False").
		Where("u.ifa_id in ", childifa).
		OrderBy("a.id asc")

	if err = q.All(&statements); err != nil {
		log.Fatal("Error fetching data from the query", err)
		return statements, err
	}

	return statements, nil
}
