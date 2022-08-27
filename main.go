package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var mailInfo MailServerInfo
var mailBodyIssued, mailBodyRecover, mailBodySSL string
var infoLogFile, errorLogFile *log.Logger
var chkRstDataHash map[int][]ChkRstData

func main() {
	var configData *ConfigData
	if len(os.Args) == 2 {
		configData = configFileScan(os.Args[1])
	} else {
		configData = configFileScan("config.yml")
	}

	if configData != nil {
		fpLogFile := logFileCheck(configData.ServerInfo.LogFile)

		if fpLogFile != nil {
			mailInfo = configData.ServerInfo.MailServerInfo

			logAddLine(infoLogFile, "Web Status Checker started")
			startTime := time.Now()

			mailBodyIssued = bodyFileScan(configData.ServerInfo.MailServerInfo.MailBodyIssuedFile, true)
			mailBodyRecover = bodyFileScan(configData.ServerInfo.MailServerInfo.MailBodyRecoverFile, false)
			mailBodySSL = bodySSLFileScan(configData.ServerInfo.MailServerInfo.MailBodySSLFile)
			if mailBodyIssued != "" && mailBodyRecover != "" {
				db := dbConn(configData.DatabaseInfo.Host, configData.DatabaseInfo.Port, configData.DatabaseInfo.Protocol, configData.DatabaseInfo.User, configData.DatabaseInfo.Password, configData.DatabaseInfo.Name)

				if db != nil && dbCheckAll(db) {
					logAddLine(infoLogFile, "Database connected")
					/*userDataList := dbQueryUserSELECT(db, "WHERE ACCOUNT='chan'")
					for _, userData := range userDataList {
						fmt.Println(userData)
					}*/

					chkRstDataHash = dbQueryChkRstSELECT(db)
					webDataList := dbQueryWebSELECT(db, "")

					if webDataList != nil {
						wg := sync.WaitGroup{}
						wg.Add(1)

						go func() {
							defer wg.Done()
							for _, webData := range webDataList {
								if time.Now().Format("15:04") == "00:00" && (strings.Split(webData.lasttime, " ")[0] == time.Now().AddDate(0, 0, -1).Format("2006-01-02") || strings.Split(webData.lasttime, " ")[0] == time.Now().Format("2006-01-02")) {
									wg.Add(1)
									go webDaysData(db, webData.idx, webData.uptimeper, &wg)
								}

								if webData.alarm != 0 {
									wg.Add(1)
									go webCheck(db, webData, configData.ServerInfo, &wg)
								}
							}
						}()

						wg.Wait()
					} else {
						logAddLine(errorLogFile, "No rows in the Web table")
					}

					if time.Now().Format("15:04") == "00:00" {
						dbQueryIU(db, "DELETE FROM WEBUPTIME WHERE CHECKDAY<DATE_ADD(NOW(), INTERVAL -90 DAY)")
					}
					dbQueryIU(db, "DELETE FROM CHKRESULT WHERE CHKTIME<DATE_ADD(NOW(), INTERVAL -24 HOUR)")
					logAddLine(infoLogFile, "Delete Older data (24H)")
				}

				elapsedTime := time.Since(startTime)
				logAddLine(infoLogFile, "Web Status Checker finished ("+strconv.FormatFloat(elapsedTime.Seconds(), 'f', 2, 64)+" sec)")

				defer fpLogFile.Close()
				dbClose(db)
			}
		}
	}
}
