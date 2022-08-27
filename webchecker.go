package main

import (
	"crypto/tls"
	"database/sql"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 웹 사이트 검사 및 메일 발송 호출
func webCheck(db *sql.DB, webData WebData, svrInfo ServerInfo, wg *sync.WaitGroup) {
	var urlIdxStr = strconv.Itoa(webData.idx)
	startTime := time.Now()
	logAddLine(infoLogFile, "IDX: "+urlIdxStr+", URL:"+webData.url+" check started")

	webChkData := urlChecker(webData.url, webData.chkcon, webData.timeout, webData.tlscheck)

	var urlStatStr = strconv.FormatBool(webChkData.urlStatus)
	var respStatStr = strconv.Itoa(webChkData.respStatus)
	var checkData = true

	if !webChkData.urlStatus || (webChkData.respStatus != 200 && webData.statcheck) {
		checkData = false
	}

	dbQueryIU(db, "INSERT INTO CHKRESULT (RESULT, STATUS, CHECKRST, URLIDX) VALUES("+urlStatStr+", "+respStatStr+", "+strconv.FormatBool(checkData)+", "+urlIdxStr+")")
	//dbQueryIU(db, "DELETE FROM CHKRESULT WHERE URLIDX="+urlIdxStr+" AND CHKTIME<DATE_ADD(NOW(), INTERVAL -24 HOUR)")
	dbQueryIU(db, "UPDATE WEB SET LASTRESULT="+urlStatStr+", LASTSTATUS="+respStatStr+", LASTTIME=NOW(), LASTCHECK="+strconv.FormatBool(checkData)+" WHERE IDX="+urlIdxStr)
	webPercent(db, webData.idx)

	if checkData && (time.Now().Format("15:04") == "00:00" && (strings.Split(webData.lasttime, " ")[0] == time.Now().AddDate(0, 0, -1).Format("2006-01-02") || strings.Split(webData.lasttime, " ")[0] == time.Now().Format("2006-01-02"))) {
		if svrInfo.SSLCheck {
			go webSSLCheck(db, webData, svrInfo.SSLCheckCycle)
		}
	}

	if webData.alarm <= 3 && !checkData {
		if webData.alarm != 3 || (webData.alarm == 3 && webData.lastcheck && !checkData) {
			mailList := strings.Split(webData.mail, " ")
			sendMail(mailList, webData.name, webData.url, webChkData.urlStatus, respStatStr, webData.chkcon, webData.rcmdtrs, webData.uptimeper, false)
		}

	} else if (webData.alarm == 2 || webData.alarm == 3) && (!webData.lastcheck && checkData) {
		mailList := strings.Split(webData.mail, " ")
		sendMail(mailList, webData.name, webData.url, webChkData.urlStatus, respStatStr, webData.chkcon, webData.rcmdtrs, webData.uptimeper, true)
	}

	elapsedTime := time.Since(startTime)
	logAddLine(infoLogFile, "IDX: "+urlIdxStr+", URL:"+webData.url+" check finished ("+strconv.FormatFloat(elapsedTime.Seconds(), 'f', 2, 64)+" sec)")
	defer wg.Done()
}

// 대상 웹 사이트 1일치 업타임 기록
func webDaysData(db *sql.DB, urlidx int, uptimeper float64, wg *sync.WaitGroup) {
	var urlIdxStr = strconv.Itoa(urlidx)
	var uptimeperStr = strconv.FormatFloat(uptimeper, 'f', 2, 64)
	dbQueryIU(db, "INSERT INTO WEBUPTIME (URLIDX, UPTIMEPER) VALUES("+urlIdxStr+", "+uptimeperStr+")")
	logAddLine(infoLogFile, "IDX: "+urlIdxStr+" Uptime recorded")
	defer wg.Done()
}

// 대상 웹 사이트의 24시간 Uptime 상태를 기록하는 함수
func webPercent(db *sql.DB, urlidx int) {
	var urlIdxStr = strconv.Itoa(urlidx)
	if chkRstDataHash != nil && chkRstDataHash[urlidx] != nil {
		var perrst float64 = 0.0

		if len(chkRstDataHash[urlidx]) == 1 {
			if chkRstDataHash[urlidx][0].check {
				perrst = 100.0
			}
		} else if len(chkRstDataHash[urlidx]) != 0 {
			dataCount := len(chkRstDataHash[urlidx])
			trueCount := 0
			for _, chkRstData := range chkRstDataHash[urlidx] {
				if chkRstData.check {
					trueCount += 1
				}
			}

			if trueCount == 0 {
				perrst = 0.0
			} else if dataCount == trueCount {
				perrst = 100.0
			} else {
				perrst = (100.0 / float64(dataCount)) * float64(trueCount)
			}
		}

		dbQueryIU(db, "UPDATE WEB SET UPTIMEPER="+strconv.FormatFloat(perrst, 'f', 2, 64)+" WHERE IDX="+urlIdxStr)
	} else {
		logAddLine(errorLogFile, "No rows in the CheckResult table")
	}
}

func webSSLCheck(db *sql.DB, webData WebData, sslCheckCycle []int) {
	if webData.tlscheck && strings.Contains(webData.url, "https") {
		u, err := url.Parse(webData.url)

		if err != nil {
			logAddLine(errorLogFile, webData.url+" URL Parse Error")
		} else {
			var domain = ""
			if host, port, err := net.SplitHostPort(u.Host); err == nil {
				domain = host + ":" + port
			} else {
				domain = u.Host + ":443"
			}

			conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 3 * time.Second}, "tcp", domain, nil)

			if err != nil {
				logAddLine(errorLogFile, webData.url+" "+domain+" doesn't support SSL Cert")
			} else {
				loc, err := time.LoadLocation("Asia/Seoul")
				if err != nil {
					logAddLine(errorLogFile, "Asia/Seoul doesn't exist.")
				} else {
					expire := conn.ConnectionState().PeerCertificates[0].NotAfter.In(loc)
					if webData.sslexpire == nil || *webData.sslexpire != expire.Format("2006-01-02 15:04:05") {
						dbQueryIU(db, "UPDATE WEB SET SSLEXPIRE='"+expire.Format("2006-01-02 15:04:05")+"' WHERE IDX="+strconv.Itoa(webData.idx))
					}

					expireDay := int(time.Until(expire).Hours() / 24)

					for _, cycleDay := range sslCheckCycle {
						if expireDay == cycleDay {
							mailList := strings.Split(webData.mail, " ")
							sendMailSSL(mailList, webData.name, webData.url, expire.Format("2006-01-02 15:04:05"))
							logAddLine(infoLogFile, webData.url+" SSL Cert Expire Date D-"+strconv.Itoa(expireDay))
						}
					}
				}
			}
		}
	} else {
		if webData.tlscheck {
			// https://가 아닌 경우에는 TLSCHECK를 비활성화 함
			dbQueryIU(db, "UPDATE WEB SET TLSCHECK=false WHERE IDX="+strconv.Itoa(webData.idx))
			logAddLine(infoLogFile, webData.url+" TLS Check replaced. (true -> false)")
		}
	}
}

// 대상 URL의 실제 접근하여 Contents를 가져오는 함수
func urlChecker(url string, checkContents string, timeo int, checkTLS bool) WebChkData {
	client := http.Client{
		Timeout: time.Duration(timeo) * time.Second,
	}

	if !checkTLS {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				Renegotiation:      tls.RenegotiateOnceAsClient,
				InsecureSkipVerify: true},
		}

		client = http.Client{
			Timeout:   10 * time.Second,
			Transport: tr,
		}
	}

	resp, err := client.Get(url)

	var webChkData WebChkData

	if err != nil {
		errCheck(err, "WEB Connect ")

		webChkData.respStatus = 0
		webChkData.urlStatus = false
		webChkData.bodyContents = ""
	} else {
		doc, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			errCheck(err, "Document Check")
			webChkData.bodyContents = ""
			webChkData.respStatus = 0
			webChkData.urlStatus = false
		} else {
			webChkData.bodyContents = string(doc)
			webChkData.respStatus = resp.StatusCode
			webChkData.urlStatus = strings.Contains(webChkData.bodyContents, checkContents)
		}

		defer resp.Body.Close()
	}

	return webChkData
}
