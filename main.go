package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const usage = `You need to set environment variables!
Example:

HILINK_API_URL="192.168.8.1" # IP address of your modem
HILINK_USER="admin" # Login and password for modem control panel
HILINK_PASS="admin"

TG_SMS_CHAN_ID="-1001205416602" # Your private Telegram channel to resend SMS to.
To get your Channel ID forward any message from your private channel to @myidbot

TG_BOT_TKN="5140291944:AA2s_Q93qVkf5KtaZ_SvKn0gwMxd0fadfd2" # Your Telegram Bot Token.
Get your token by creating bot at https://t.me/BotFather Invite your bot to your private channel and give him "Post messages" permission.

If you have Docker installed you can just run one command: 

docker run -d -e HILINK_API_URL='192.168.8.1' -e HILINK_USER='admin' -e HILINK_PASS='admin' -e TG_SMS_CHAN_ID='-1001395206691' -e TG_BOT_TKN='1330281164:FFEs_SzdqGk521tVZ_VtKn04wMwPzfKGjgG' --name hlsms wmw9/hilink-sms-resender && docker logs hlsms -f

`

var (
	HilinkApiUrl = fmt.Sprintf("http://%v", os.Getenv("HILINK_API_URL")) // Your modem IP address
	HilinkUser   = os.Getenv("HILINK_USER")
	HilinkPass   = os.Getenv("HILINK_PASS")
	TgUrl        = fmt.Sprintf("https://api.telegram.org/bot%v/sendMessage", os.Getenv("TG_BOT_TKN"))
	tgSMSChanID  = os.Getenv("TG_SMS_CHAN_ID")
)

type SesTokInfo struct {
	SesInfo string `xml:"SesInfo"`
	TokInfo string `xml:"TokInfo"`
}

type SmsResponse struct {
	Count    string   `xml:"Count"`
	Messages struct {
		Message struct {
			Index   string `xml:"Index"`
			Phone   string `xml:"Phone"`
			Content string `xml:"Content"`
			Date    string `xml:"Date"`
		} `xml:"Message"`
	} `xml:"Messages"`
}

// Message is a Telegram object that can be found in an update.
type Message struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

var sestokinfo SesTokInfo
var smsresponse SmsResponse

func main() {
	checkEnvs()
	fmt.Printf("[*] HiLink SMS Resender started. \nHost: %v\nUser: %v\nPass: %v\n\n",
		HilinkApiUrl, HilinkUser, HilinkPass)
	login()
	for {
		if new := getSms(); new {
			if ok := resendSms(); ok {
				log.Println("SMS sent!")
				deleteSms()
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func checkEnvs() {
	if HilinkApiUrl == "http://" ||
		HilinkPass == "" ||
		HilinkUser == "" ||
		tgSMSChanID == "" ||
		TgUrl == "https://api.telegram.org/bot/sendMessage" {
		fmt.Println(usage)
		os.Exit(3)
	}
}

// createRequest is creating NewRequest with POST/GET method
func createRequest(url string, body ...string) (*http.Request, error) {
	var req *http.Request
	if body == nil {
		req, _ = http.NewRequest("GET", url, nil)
	} else {
		data := strings.NewReader(body[0])
		req, _ = http.NewRequest("POST", url, data)
	}

	req.Header.Set("Cookie", sestokinfo.SesInfo)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("__RequestVerificationToken", sestokinfo.TokInfo)

	return req, nil
}

func createClient() *http.Client {
	return &http.Client{Timeout: 300 * time.Second}
}

func reqDo(c *http.Client, req *http.Request) ([]byte, *http.Response, error) {
	res, err := c.Do(req)
	if err != nil {
		tgSendError(err)
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		tgSendError(err)
		panic(err)
	}
	//fmt.Println(res.Status)

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		tgSendError(err)
		panic(err)
	}
	return bytes, res, nil
}

// getSesTokinfo is generating new token and sessionID by making request to HiLink API "/api/webserver/SesTokInfo"
func getSesTokInfo() {
	url := fmt.Sprintf("%v/api/webserver/SesTokInfo", HilinkApiUrl)
	//fmt.Printf("%v -> ", url)

	req, err := createRequest(url)
	if err != nil {
		panic(err)
	}
	c := createClient()

	xmlBytes, _, _ := reqDo(c, req)
	xml.Unmarshal(xmlBytes, &sestokinfo)

	//fmt.Println("body:", string(xmlBytes))
	//fmt.Printf("SesInfo: %v\nTokInfo: %v\n", sestokinfo.SesInfo, sestokinfo.TokInfo)
}

// login is used to authorize us using user and password
func login() {
	getSesTokInfo()
	tokenizedPw := hashPw(HilinkUser + hashPw(HilinkPass) + sestokinfo.TokInfo)

	data := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?><request><Username>%v</Username><Password>%v</Password><password_type>4</password_type></request>`, HilinkUser, tokenizedPw)

	url := fmt.Sprintf("%v/api/user/login", HilinkApiUrl)
	//fmt.Println("url:", url, "SesInfo:", sestokinfo.SesInfo)

	req, err := createRequest(url, data)
	if err != nil {
		panic(err)
	}
	c := createClient()

	_, res, _ := reqDo(c, req)

	//fmt.Println("body:", string(body))

	// Saving cookie from response
	sestokinfo.SesInfo = strings.Split(res.Header.Get("Set-Cookie"), ";")[0]
	//fmt.Println("cookie:", sestokinfo.SesInfo)
}

func getSms() bool {
	getSesTokInfo()
	data := `<?xml version="1.0" encoding="UTF-8"?><request><PageIndex>1</PageIndex><ReadCount>1</ReadCount><BoxType>1</BoxType><SortType>0</SortType><Ascending>0</Ascending><UnreadPreferred>0</UnreadPreferred></request>`

	url := fmt.Sprintf("%v/api/sms/sms-list", HilinkApiUrl)

	req, err := createRequest(url, data)
	if err != nil {
		panic(err)
	}
	c := createClient()

	xmlBytes, _, _ := reqDo(c, req)
	smsresponse = SmsResponse{}
	xml.Unmarshal(xmlBytes, &smsresponse)
	if smsresponse.Count == "0" {
		return false
	}
	log.Printf("SMS index: %v\nFrom: %v\nText: %v\nDate: %v\n", smsresponse.Messages.Message.Index,
		smsresponse.Messages.Message.Phone,
		smsresponse.Messages.Message.Content,
		smsresponse.Messages.Message.Date)
	return true
}

func deleteSms() {
	getSesTokInfo()
	data := fmt.Sprintf(`<?xml version=\"1.0\" encoding=\"UTF-8\"?><request><Index>%v</Index></request>`,
		smsresponse.Messages.Message.Index)
	//fmt.Println("data:", data)

	url := fmt.Sprintf("%v/api/sms/delete-sms", HilinkApiUrl)
	//fmt.Println("url:", url, "SesInfo:", sestokinfo.SesInfo)

	req, err := createRequest(url, data)
	if err != nil {
		panic(err)
	}
	c := createClient()

	_, _, _ = reqDo(c, req)
}

func resendSms() bool {
	phone := smsresponse.Messages.Message.Phone
	text := smsresponse.Messages.Message.Content
	date := smsresponse.Messages.Message.Date
	message := Message{}
	message.ChatID = tgSMSChanID
	message.ParseMode = "HTML"
	message.Text = fmt.Sprintf("<b>%v:</b> %v\n\n<pre>%v</pre>", phone, text, date)

	js, _ := json.Marshal(message)
	res, err := http.Post(TgUrl, "application/json", bytes.NewBuffer(js))
	if err != nil {
		tgSendError(err)
		panic(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		tgSendError(err)
		return false
	}
	return true

}

func hashPw(text string) string {
	h := sha256.New()
	h.Write([]byte(text))
	hash := hex.EncodeToString(h.Sum(nil))
	return base64.StdEncoding.EncodeToString([]byte(hash))
}

func tgSendError(e interface{}) {
	log.Printf("%v", e)
	message := Message{}
	message.ChatID = tgSMSChanID
	message.Text = fmt.Sprintf("%v", e)

	js, _ := json.Marshal(message)
	res, err := http.Post(TgUrl, "application/json", bytes.NewBuffer(js))
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
}
