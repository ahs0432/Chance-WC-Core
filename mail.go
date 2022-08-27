package main

import (
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// TLS 인증 오류 해소를 위한 Auth struct를 생성
type unencrptedAuth struct {
	smtp.Auth
}

// TLS 인증 오류 해소를 위해 a.Auth 함수 상 추가될 일부 데이터를 변조하여 추가
func (a unencrptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}

// 메일 발송
func sendMail(mailTo []string, name string, url string, urlStat bool, respStat string, chkcon string, rcmdtrs string, uptimeper float64, alarmStat bool) {
	mailToStr := "To: "

	for _, mailToTemp := range mailTo {
		mailToStr = mailToStr + ", " + mailToTemp
	}

	urlStatStr := ""

	if urlStatStr = "콘텐츠 확인 실패"; urlStat {
		urlStatStr = "콘텐츠 확인 성공"
	}

	if respStat == "0" {
		respStat = "사이트 연결 실패"
	}

	var subject = "Subject: "

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	var body string
	if alarmStat {
		body = mailBodyRecover
		subject = subject + strings.ReplaceAll(strings.ReplaceAll(mailInfo.MailSubjectRecover, "(-URL-)", url), "(-NAME-)", name) + "\r\n"
	} else {
		body = mailBodyIssued
		subject = subject + strings.ReplaceAll(strings.ReplaceAll(mailInfo.MailSubjectIssued, "(-URL-)", url), "(-NAME-)", name) + "\r\n"
	}

	body = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		body, "(-URL-)", url),
		"(-TIME-)", time.Now().Format("2006-01-02(Mon) 15:04:05")),
		"(-STATUSCODE-)", respStat),
		"(-CHECKCONTENTS-)", chkcon),
		"(-CHECKRESULT-)", urlStatStr),
		"(-DAYUPTIME-)", strconv.FormatFloat(uptimeper, 'f', 2, 64)),
		"(-RECOMMEND-)", rcmdtrs),
		"(-NAME-)", name)

	msg := []byte("From: Web Site Monitor <" + mailInfo.MailFrom + ">\r\n" + mailToStr + "\r\n" + subject + "Date:" + time.Now().Format(time.RFC1123Z) + "\r\n" + mime + "\r\n" + body)
	auth := unencrptedAuth{
		smtp.PlainAuth("", mailInfo.UserName, mailInfo.MailPassword, mailInfo.SMTPHost),
	}
	err := smtp.SendMail(mailInfo.SMTPHost+":"+mailInfo.SMTPPort, auth, mailInfo.MailFrom, mailTo, msg)
	if err != nil {
		errCheck(err, "SendMail")
	}
}

// SSL 만료일 안내 메일 발송
func sendMailSSL(mailTo []string, name string, url string, expireDate string) {
	mailToStr := "To: "

	for _, mailToTemp := range mailTo {
		mailToStr = mailToStr + ", " + mailToTemp
	}

	var subject = "Subject: "

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := mailBodySSL
	subject = subject + strings.ReplaceAll(strings.ReplaceAll(mailInfo.MailSubjectSSL, "(-URL-)", url), "(-NAME-)", name) + "\r\n"

	body = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		body, "(-URL-)", url),
		"(-SSLTIME-)", expireDate),
		"(-NAME-)", name)

	msg := []byte("From: Web Site Monitor <" + mailInfo.MailFrom + ">\r\n" + mailToStr + "\r\n" + subject + "Date:" + time.Now().Format(time.RFC1123Z) + "\r\n" + mime + "\r\n" + body)
	auth := unencrptedAuth{
		smtp.PlainAuth("", mailInfo.UserName, mailInfo.MailPassword, mailInfo.SMTPHost),
	}
	err := smtp.SendMail(mailInfo.SMTPHost+":"+mailInfo.SMTPPort, auth, mailInfo.MailFrom, mailTo, msg)
	if err != nil {
		errCheck(err, "SendMail")
	}
}
