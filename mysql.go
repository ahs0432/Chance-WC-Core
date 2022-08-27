package main

import (
	"database/sql"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB 연결 함수
func dbConn(host string, port string, protocol string, user string, password string, dbname string) *sql.DB {
	db, err := sql.Open("mysql", user+":"+password+"@"+protocol+"("+host+":"+port+")/"+dbname)

	if err != nil {
		errCheck(err, "Database Connection")
	} else {
		db.SetConnMaxLifetime(time.Minute * 5)
		db.SetConnMaxIdleTime(time.Minute * 1)
		db.SetMaxOpenConns(1024)
		db.SetMaxIdleConns(1024)

		return db
	}
	return nil
}

// DB 종료
func dbClose(db *sql.DB) {
	db.Close()
}

// Table 존재 유무 확인 함수
func dbCheckTable(db *sql.DB, table string) bool {
	_, table_check := db.Query("SELECT * FROM " + table)

	if table_check != nil {
		logAddLine(infoLogFile, "Database "+table+" not found")
	}

	return table_check == nil
}

// Table 생성 사용 함수
func dbCreateTable(db *sql.DB, query string) {
	_, err := db.Exec(query)
	if err != nil {
		errCheck(err, "Database Create Table")
	}
}

/*// User 데이터 검색 함수 (WHERE 절의 경우 별도 입력)
func dbQueryUserSELECT(db *sql.DB, where string) []UserData {
	queryRow, err := db.Query("SELECT IDX, ACCOUNT, AES_DECRYPT(unhex(PASSWORD), 'chanceen'), MAIL FROM USERS " + where)
	if err != nil {
		errCheck(err, "Database User Select Query")
	}
	defer queryRow.Close()

	var userDataList []UserData
	var userData UserData

	for queryRow.Next() {
		err := queryRow.Scan(&userData.idx, &userData.id, &userData.passwd, &userData.mail)
		if err != nil {
			userDataList = nil
			errCheck(err, "Database User Select Query Scan")
		} else {
			if userDataList != nil {
				userDataList = append(userDataList, userData)
			} else {
				userDataList = []UserData{userData}
			}
		}
	}

	return userDataList
}*/

// Web 테이블 데이터 검색 함수 (WHERE 절의 경우 별도 입력)
func dbQueryWebSELECT(db *sql.DB, where string) []WebData {
	queryRow, err := db.Query("SELECT * FROM WEB " + where)
	if err != nil {
		errCheck(err, "Database Web Select Query")
	}

	defer queryRow.Close()

	var webDataList []WebData
	var webData WebData

	for queryRow.Next() {
		err := queryRow.Scan(&webData.idx, &webData.name, &webData.url, &webData.chkcon, &webData.rcmdtrs, &webData.mail, &webData.lastresult, &webData.laststatus, &webData.lastcheck, &webData.lasttime, &webData.sslexpire, &webData.uptimeper, &webData.tlscheck, &webData.statcheck, &webData.alarm, &webData.timeout, &webData.useridx)
		if err != nil {
			webDataList = nil
			errCheck(err, "Database Web Select Query Scan")
		} else {
			if webDataList != nil {
				webDataList = append(webDataList, webData)
			} else {
				webDataList = []WebData{webData}
			}
		}
	}

	return webDataList
}

// CHKRESULT 테이블 데이터 검색 함수
func dbQueryChkRstSELECT(db *sql.DB) map[int][]ChkRstData {
	queryRow, err := db.Query("SELECT * FROM CHKRESULT")

	chkRstHash := map[int][]ChkRstData{}

	if err != nil {
		errCheck(err, "Database CheckResult Select Query")
	} else {
		defer queryRow.Close()

		for queryRow.Next() {
			var chkRstDataList []ChkRstData
			var chkRstData ChkRstData

			err := queryRow.Scan(&chkRstData.result, &chkRstData.status, &chkRstData.check, &chkRstData.chktime, &chkRstData.urlidx)

			if err != nil {
				errCheck(err, "Database Web Select Query Scan")
			} else {
				_, exist := chkRstHash[chkRstData.urlidx]

				if exist {
					chkRstDataList = chkRstHash[chkRstData.urlidx]
					chkRstDataList = append(chkRstDataList, chkRstData)
					chkRstHash[chkRstData.urlidx] = chkRstDataList
				} else {
					chkRstDataList = []ChkRstData{chkRstData}
					chkRstHash[chkRstData.urlidx] = chkRstDataList
				}
			}
		}

		logAddLine(infoLogFile, "Check Result Data (Uptime) SELECT Complete")
	}

	return chkRstHash
}

// 데이터 추가/변경 함수
func dbQueryIU(db *sql.DB, query string) {
	queryRow, err := db.Exec(query)
	if err != nil {
		errCheck(err, "Database Insert/Update Query")
	} else {
		_, err = queryRow.RowsAffected()
		//countQuery, err := queryRow.RowsAffected()
		if err != nil {
			errCheck(err, "Database Insert/Update RowAffected")
		}
	}
}

// DB 검사 및 테이블 생성 진행 함수
func dbCheckAll(db *sql.DB) bool {
	err := db.Ping()
	if err != nil {
		errCheck(err, "Database Ping Check")
		return false
	}

	if !dbCheckTable(db, "USERS") {
		dbCreateTable(db, `CREATE TABLE USERS (
							IDX INT(20) AUTO_INCREMENT PRIMARY KEY,
							ACCOUNT VARCHAR(20) NOT NULL UNIQUE,
							PASSWORD VARCHAR(50) NOT NULL,
							MAIL VARCHAR(50) NOT NULL
							);`)
		logAddLine(infoLogFile, "Database USERS table created")
	}

	if !dbCheckTable(db, "WEB") {
		dbCreateTable(db, `CREATE TABLE WEB (
							IDX BIGINT(30) AUTO_INCREMENT PRIMARY KEY,
							NAME VARCHAR(30) NOT NULL,
							URL LONGTEXT NOT NULL,
							CHKCON VARCHAR(100) NOT NULL DEFAULT 'TEST',
							RCMDTRS LONGTEXT NOT NULL DEFAULT '',
							MAIL LONGTEXT NOT NULL,
							LASTRESULT BOOLEAN NOT NULL DEFAULT FALSE,
							LASTSTATUS INT(3) NOT NULL DEFAULT 0,
							LASTCHECK BOOLEAN NOT NULL DEFAULT TRUE,
							LASTTIME DATETIME NOT NULL DEFAULT NOW(),
							SSLEXPIRE DATETIME,
							UPTIMEPER FLOAT(5) NOT NULL DEFAULT 0.0,
							TLSCHECK BOOLEAN NOT NULL DEFAULT TRUE,
							STATCHECK BOOLEAN NOT NULL DEFAULT TRUE,
							ALARM INT(1) NOT NULL DEFAULT 1,
							TIMEOUT INT(2) NOT NULL DEFAULT 10,
							USERIDX INT(20)
							);`)
		logAddLine(infoLogFile, "Database WEB table created")
	}

	if !dbCheckTable(db, "CHKRESULT") {
		dbCreateTable(db, `CREATE TABLE CHKRESULT (
							RESULT BOOLEAN NOT NULL,
							STATUS INT(3) NOT NULL,
							CHECKRST BOOLEAN NOT NULL,
							CHKTIME DATETIME NOT NULL DEFAULT NOW(),
							URLIDX BIGINT(30)
							);`)
		logAddLine(infoLogFile, "Database CHKRESULT table created")
	}

	if !dbCheckTable(db, "WEBUPTIME") {
		dbCreateTable(db, `CREATE TABLE WEBUPTIME (
							URLIDX BIGINT(30) NOT NULL,
							UPTIMEPER FLOAT(5) NOT NULL,
							CHECKDAY DATE NOT NULL DEFAULT DATE_ADD(NOW(), INTERVAL -1 DAY)
							);`)
		logAddLine(infoLogFile, "Database WEBUPTIME table created")
	}

	if !dbCheckTable(db, "WEBGROUP") {
		dbCreateTable(db, `CREATE TABLE WEBGROUP (
							NAME VARCHAR(100) NOT NULL UNIQUE,
							MEMBER LONGTEXT NOT NULL DEFAULT '',
							COUNT BIGINT(30) NOT NULL DEFAULT 0
							);`)
		logAddLine(infoLogFile, "Database WEBGROUP table created")
	}

	return true
}
