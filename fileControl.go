package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// 각각의 로그를 기입할 수 있도록 로그 대상 형태와 내용을 수신하여 작성하는 함수
func logAddLine(addLogFile *log.Logger, logContents string) {
	addLogFile.Print(logContents)
}

// 로그 파일에 이상이 없는지 확인하고 각 로그 파일을 할당하는 함수
func logFileCheck(logFileName string) *os.File {
	fpLog, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		infoLogFile = log.New(fpLog, "INFO: ", log.Ltime|log.Ldate)
		errorLogFile = log.New(fpLog, "ERROR: ", log.Ltime|log.Ldate)
		return fpLog
	}
	return nil
}

func configFileCreate(confFile string) bool {
	mailServerInfo := MailServerInfo{
		UserName:            "webchecker",
		MailFrom:            "webchecker@test.com",
		MailPassword:        "P@ssW0rd",
		SMTPHost:            "mail.test.com",
		SMTPPort:            "25",
		MailSubjectSSL:      "[엔클라우드24] (-NAME-) 사이트 인증서 만료 예정 - (-URL-)",
		MailSubjectIssued:   "[엔클라우드24] (-NAME-) 사이트 알람 발생 - (-URL-)",
		MailSubjectRecover:  "[엔클라우드24] (-NAME-) 사이트 알람 해소 - (-URL-)",
		MailBodySSLFile:     "./_html/mailBodySSL.html",
		MailBodyIssuedFile:  "./_html/mailBodyIssued.html",
		MailBodyRecoverFile: "./_html/mailBodyRecover.html"}

	databaseInfo := DatabaseInfo{
		Host:     "127.0.0.1",
		Port:     "3306",
		Protocol: "tcp",
		User:     "webchecker",
		Password: "P@ssW0rd",
		Name:     "webchecker"}

	serverInfo := ServerInfo{
		LogFile:        "./WebChecker.log",
		SSLCheck:       true,
		SSLCheckCycle:  []int{1, 3, 5, 15, 30},
		MailServerInfo: mailServerInfo}

	configData := ConfigData{
		DatabaseInfo: databaseInfo,
		ServerInfo:   serverInfo,
	}
	data, err := yaml.Marshal(&configData)

	if err != nil {
		fmt.Println(confFile + " file create failed")
		fmt.Println(err)
		return false
	}

	err2 := ioutil.WriteFile(confFile, data, 0644)
	if err2 != nil {
		fmt.Println(confFile + " file create failed")
		fmt.Println(err)
		return false
	}

	fmt.Println(confFile + " file create")
	return true
}

func configFileScan(confFile string) *ConfigData {
	filename, _ := filepath.Abs(confFile)
	yamlFile, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Println(err)
		if strings.Contains(err.Error(), "The system cannot find the file specified.") || strings.Contains(err.Error(), "no such file or directory") {
			if !configFileCreate(confFile) {
				return nil
			}

			return configFileScan(confFile)
		} else {
			return nil
		}
	}

	configData := &ConfigData{}
	err = yaml.Unmarshal([]byte(yamlFile), &configData)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	return configData
}

func bodyFileScan(bodyFileName string, issuedType bool) string {
	filename, _ := filepath.Abs(bodyFileName)
	mailBody, err := ioutil.ReadFile(filename)

	if err != nil {
		errCheck(err, "Body contents check")
		if strings.Contains(err.Error(), "The system cannot find") || strings.Contains(err.Error(), "no such file or directory") {
			// folder check
			folders := strings.Split(bodyFileName, "/")
			folder := ""
			for i := 0; i < len(folders)-1; i++ {
				folder += folders[i] + "/"
			}

			foldername, _ := filepath.Abs(folder)
			_, errDir := ioutil.ReadDir(foldername)

			if errDir != nil {
				errCheck(errDir, "No such Directory.")
				logAddLine(infoLogFile, folder+" folder created")
				os.MkdirAll(foldername, os.ModePerm)
			}

			defaultHTML := `
			<body>
				<table align="center" width="60%" style="border-collapse: collapse; border-radius: 8px">
					<th align="left" colspan="2"><h2>Chance WC</h2></th>
					<tr><td height="25"><td></tr>
					<tr align="center">
						<td colspan="2" style="font-size: 25px"><b>사이트에서 이상이 감지되었습니다!</b></td>
					</tr>
					<tr><td height="40"><td></tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>사이트 명</b></td>
						<td>(-NAME-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>사이트 주소</b></td>
						<td>(-URL-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>발생 시간</b></td>
						<td>(-TIME-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>상태 코드</b></td>
						<td>(-STATUSCODE-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>감지 콘텐츠</b></td>
						<td>(-CHECKCONTENTS-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>콘텐츠 체크</b></td>
						<td>(-CHECKRESULT-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>최근 24시간 이내 업타임</b></td>
						<td>(-DAYUPTIME-) %</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>사이트 권장 해소 방안</b></td>
						<td>(-RECOMMEND-)</td>
					</tr>
				</table>
			</body>`

			if !issuedType {
				defaultHTML = strings.Replace(defaultHTML, "사이트에서 이상이 감지되었습니다", "사이트 이상이 해소되었습니다", 1)
			}

			err := ioutil.WriteFile(bodyFileName, []byte(defaultHTML), 0644)

			if err != nil {
				errCheck(err, "Body contents check")
				return ""
			}

			logAddLine(infoLogFile, bodyFileName+" file created")
			return defaultHTML
		} else {
			return ""
		}
	}

	return string(mailBody)
}

func bodySSLFileScan(bodyFileName string) string {
	filename, _ := filepath.Abs(bodyFileName)
	mailBody, err := ioutil.ReadFile(filename)

	if err != nil {
		errCheck(err, "Body contents check")
		if strings.Contains(err.Error(), "The system cannot find the file specified.") || strings.Contains(err.Error(), "no such file or directory") {
			// folder check
			folders := strings.Split(bodyFileName, "/")
			folder := ""
			for i := 0; i < len(folders)-1; i++ {
				folder += folders[i] + "/"
			}

			foldername, _ := filepath.Abs(folder)
			_, errDir := ioutil.ReadDir(foldername)

			if errDir != nil {
				errCheck(errDir, "No such Directory.")
				logAddLine(infoLogFile, folder+" folder created")
				os.MkdirAll(foldername, os.ModePerm)
			}

			defaultHTML := `
			<body>
				<table align="center" width="60%" style="border-collapse: collapse; border-radius: 8px">
					<th align="left" colspan="2"><h2>Chance WC</h2></th>
					<tr><td height="25"><td></tr>
					<tr align="center">
						<td colspan="2" style="font-size: 25px"><b>SSL 인증서 만료일이 도래되었습니다.</b></td>
					</tr>
					<tr><td height="40"><td></tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>사이트 명</b></td>
						<td>(-NAME-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>사이트 주소</b></td>
						<td>(-URL-)</td>
					</tr>
					<tr align="center" style="font-size: 12px">
						<td width="40%" bgcolor='#f6f6f8'><b>만료 예정 일정</b></td>
						<td>(-SSLTIME-)</td>
					</tr>
				</table>
			</body>`

			err := ioutil.WriteFile(bodyFileName, []byte(defaultHTML), 0644)

			if err != nil {
				errCheck(err, "Body contents check")
				return ""
			}

			logAddLine(infoLogFile, bodyFileName+" file created")
			return defaultHTML
		} else {
			return ""
		}
	}

	return string(mailBody)
}
