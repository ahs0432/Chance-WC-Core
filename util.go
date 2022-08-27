package main

// Error Log를 위한 함수
func errCheck(err error, errLoc string) {
	logAddLine(errorLogFile, errLoc+" Error! / "+err.Error())
}
