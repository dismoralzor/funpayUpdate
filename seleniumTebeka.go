package SeleniumTebeka

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

func FunpayUpdate(username string, password string) {

	const (
		chromeDriverPath = "./cmd/chromedriver"
		port             = 5050
	)
	opts := []selenium.ServiceOption{
		selenium.ChromeDriver(chromeDriverPath),
		selenium.Output(os.Stderr),
	}
	selenium.SetDebug(true)
	service, err := selenium.NewChromeDriverService(chromeDriverPath, port, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{
		"acceptInsecureCerts": true,
		"browserName":         "chrome",
	}
	caps.AddChrome(chrome.Capabilities{
		Args: []string{
			"no-sandbox",
			"ash-host-window-bounds", "1024x768",
			"headless",
			"disable-ios-password-suggestions",
			"allow-cross-origin-auth-prompt",
		},
		W3C:             true,
		ExcludeSwitches: []string{"enable-automation"},
		Prefs: map[string]interface{}{
			"credentials_enable_service": false,
			"password_manager_enabled":   false},
	})

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Fatal(err)
	}
	defer wd.Quit()

	if err := wd.Get("https://funpay.com/account/login?gate=vk"); err != nil {
		log.Fatal(err)
	}

	login, err := wd.FindElement(selenium.ByXPATH, "//*[@id='login_submit']/div/div/input[6]")
	if err != nil {
		log.Fatal(err)
	}

	err = login.SendKeys(username)

	pass, err := wd.FindElement(selenium.ByXPATH, "//*[@id='login_submit']/div/div/input[7]")
	if err != nil {
		log.Fatal(err)
	}

	err = pass.SendKeys(password)

	btn, err := wd.FindElement(selenium.ByXPATH, "//*[@id='install_allow']")
	if err != nil {
		log.Fatal(err)
	}
	if err := btn.Click(); err != nil {
		log.Fatal(err)
	}

	if err := wd.Get("https://funpay.com/lots/1120/trade"); err != nil {
		log.Fatal(err)
	}

	up, err := wd.FindElement(selenium.ByXPATH, "//*[@id='content']/div/div/div[2]/div/div[1]/div[2]/div/div[1]/button")
	if err != nil {
		log.Fatal(err)
	}
	if err := up.Click(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 1)
	screen, err := wd.Screenshot()
	if err != nil {
		log.Fatal(err)
	}
	ioutil.WriteFile("Screenshot.jpg", screen, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
