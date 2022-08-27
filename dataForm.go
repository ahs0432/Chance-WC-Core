package main

/*type UserData struct {
	idx    int
	id     string
	passwd string
	mail   string
}*/

type WebData struct {
	idx        int
	name       string
	url        string
	chkcon     string
	rcmdtrs    string
	mail       string
	lastresult bool
	laststatus int
	lastcheck  bool
	lasttime   string
	sslexpire  *string
	uptimeper  float64
	tlscheck   bool
	statcheck  bool
	alarm      int
	timeout    int
	useridx    int
}

type ChkRstData struct {
	result  bool
	status  int
	check   bool
	chktime string
	urlidx  int
}

type WebChkData struct {
	respStatus   int
	urlStatus    bool
	bodyContents string
}

type ConfigData struct {
	DatabaseInfo DatabaseInfo `yaml:"DatabaseInfo"`
	ServerInfo   ServerInfo   `yaml:"ServerInfo"`
}

type DatabaseInfo struct {
	Host     string `yaml:"Host"`
	Port     string `yaml:"Port"`
	Protocol string `yaml:"Protocol"`
	User     string `yaml:"User"`
	Password string `yaml:"Password"`
	Name     string `yaml:"Name"`
}

type MailServerInfo struct {
	UserName            string `yaml:"UserName"`
	MailFrom            string `yaml:"MailFrom"`
	MailPassword        string `yaml:"MailPassword"`
	SMTPHost            string `yaml:"SMTPHost"`
	SMTPPort            string `yaml:"SMTPPort"`
	MailSubjectSSL      string `yaml:"MailSubjectSSL"`
	MailSubjectIssued   string `yaml:"MailSubjectIssued"`
	MailSubjectRecover  string `yaml:"MailSubjectRecover"`
	MailBodySSLFile     string `yaml:"MailBodySSLFile"`
	MailBodyIssuedFile  string `yaml:"MailBodyIssuedFile"`
	MailBodyRecoverFile string `yaml:"MailBodyRecoverFile"`
}

type ServerInfo struct {
	LogFile        string         `yaml:"LogFile"`
	SSLCheck       bool           `yaml:"SSLCheck"`
	SSLCheckCycle  []int          `yaml:"SSLCheckCycle"`
	MailServerInfo MailServerInfo `yaml:"MailServerInfo"`
}
