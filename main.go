package main

import (
	"bufio"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/fatih/color"
	resty "gopkg.in/resty.v1"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DataName dataname
}

type dataname struct {
	SMTP_HOST string
	SMTP_PORT int
	EMAIL     string
	PASSWORD  string
	SEND_TO   string
	SEND_CC1  string
	SEND_CC2  string
}

func main() {
	mainProgram()
}

func getClient(service string, number_loop int) {
	res, err := resty.R().Get(service)
	a := strconv.Itoa(int(number_loop))
	b := strconv.Itoa(int(res.StatusCode()))
	if err == nil {
		if res.StatusCode() == 200 {
			fmt.Println(number_loop, "status code", res.StatusCode(), " on service : ", service)
		} else {
			color.Red(a + " error endpoint with status code : " + b + " on service : " + service)
			Mail(service, res.StatusCode(), "")
		}
	} else {
		color.Red(a + " endpoint failur : " + err.Error())
		Mail(service, res.StatusCode(), err.Error())
	}
}

func mainProgram() {
	for true {
		time.Sleep(2 * time.Second)
		file, err := os.Open("/etc/monitor/apps.conf")
		if err != nil {
			log.Fatalf("failed opening file: apps.conf at /etc/monitor directory")
		}
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)
		var txtlines []string

		for scanner.Scan() {
			txtlines = append(txtlines, scanner.Text())
		}

		file.Close()
		for num_app, eachline_name_of_app := range txtlines {
			getClient(eachline_name_of_app, num_app)
			time.Sleep(1 * time.Second)
		}
	}
}

func Mail(s string, status_code int, res string) {
	var conf Config
	stcode := strconv.Itoa(status_code)
	if _, err := toml.DecodeFile("/etc/monitor/config.toml", &conf); err != nil {
		fmt.Println(err)
	}
	to := []string{conf.DataName.SEND_TO, conf.DataName.SEND_TO}
	/*CC ?*/
	cc := []string{conf.DataName.SEND_CC1, conf.DataName.SEND_CC2}
	subject := "Service " + s + " is Offline"
	message := "Service " + s + " is offline" + "\nOffline time is " + getTime().String() + "\nStatus Code : " + stcode + "\nLog Error : " + res
	err := sendMail(to, cc, subject, message)
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Println("Mail sent!")
}
func sendMail(to []string, cc []string, subject, message string) error {
	var conf Config
	if _, err := toml.DecodeFile("/etc/monitor/config.toml", &conf); err != nil {
		fmt.Println(err)
	}
	fmt.Println("running .... ")
	body := "From: " + conf.DataName.EMAIL + "\n" +
		"To: " + strings.Join(to, ",") + "\n" +
		"Cc: " + strings.Join(cc, ",") + "\n" +
		"Subject: " + subject + "\n\n" +
		message
	auth := smtp.PlainAuth("", conf.DataName.EMAIL, conf.DataName.PASSWORD, conf.DataName.SMTP_HOST)
	smtpAddr := fmt.Sprintf("%s:%d", conf.DataName.SMTP_HOST, conf.DataName.SMTP_PORT)
	fmt.Println("check ", smtpAddr)
	err := smtp.SendMail(smtpAddr, auth, conf.DataName.EMAIL, append(to, cc...), []byte(body))
	if err != nil {
		return err
	}
	return nil
}
func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
	}
	return hostname
}
func getTime() time.Time {
	return time.Now()
}
