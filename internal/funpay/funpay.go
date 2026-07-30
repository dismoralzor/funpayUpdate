// Package funpay bumps ("поднимает") FunPay lots by driving a headless
// Chrome session that authenticates through the VK login gate.
package funpay

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

const (
	loginURL = "https://funpay.com/account/login?gate=vk"

	usernameXPath = "//*[@id='login_submit']/div/div/input[6]"
	passwordXPath = "//*[@id='login_submit']/div/div/input[7]"
	allowXPath    = "//*[@id='install_allow']"
	bumpXPath     = "//*[@id='content']/div/div/div[2]/div/div[1]/div[2]/div/div[1]/button"

	// renderDelay gives the page a moment to repaint after the bump click so
	// the screenshot shows the result rather than the pre-click state.
	renderDelay = time.Second
)

// Config describes a single bump run.
type Config struct {
	// Username and Password are the VK credentials used at the login gate.
	Username string
	Password string

	// ChromeDriverPath is the chromedriver executable. When empty it is
	// resolved from PATH.
	ChromeDriverPath string

	// Port is the local port chromedriver listens on.
	Port int

	// LotsURL is the lot category page holding the bump button.
	LotsURL string

	// ScreenshotPath, when non-empty, receives a PNG of the page after the
	// bump so a failed run can be inspected after the fact.
	ScreenshotPath string
}

// Validate reports whether the config is usable, filling in defaults for the
// optional fields.
func (c *Config) Validate() error {
	if c.Username == "" {
		return errors.New("username is empty")
	}
	if c.Password == "" {
		return errors.New("password is empty")
	}
	if c.LotsURL == "" {
		return errors.New("lots URL is empty")
	}
	if c.Port == 0 {
		c.Port = 5050
	}
	if c.ChromeDriverPath == "" {
		path, err := exec.LookPath("chromedriver")
		if err != nil {
			return fmt.Errorf("locate chromedriver: %w", err)
		}
		c.ChromeDriverPath = path
	}
	return nil
}

// Update performs one login-and-bump cycle. It returns as soon as any step
// fails; the caller decides whether to retry.
func Update(ctx context.Context, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	service, err := selenium.NewChromeDriverService(
		cfg.ChromeDriverPath,
		cfg.Port,
		selenium.Output(os.Stderr),
	)
	if err != nil {
		return fmt.Errorf("start chromedriver: %w", err)
	}
	defer service.Stop()

	wd, err := selenium.NewRemote(capabilities(), fmt.Sprintf("http://localhost:%d/wd/hub", cfg.Port))
	if err != nil {
		return fmt.Errorf("connect to chromedriver: %w", err)
	}
	defer wd.Quit()

	if err := wd.Get(loginURL); err != nil {
		return fmt.Errorf("open login page: %w", err)
	}
	if err := sendKeys(wd, usernameXPath, cfg.Username, "username field"); err != nil {
		return err
	}
	if err := sendKeys(wd, passwordXPath, cfg.Password, "password field"); err != nil {
		return err
	}
	if err := click(wd, allowXPath, "authorize button"); err != nil {
		return err
	}

	if err := wd.Get(cfg.LotsURL); err != nil {
		return fmt.Errorf("open lots page: %w", err)
	}
	if err := click(wd, bumpXPath, "bump button"); err != nil {
		return err
	}

	if cfg.ScreenshotPath == "" {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(renderDelay):
	}
	return screenshot(wd, cfg.ScreenshotPath)
}

func capabilities() selenium.Capabilities {
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
		// Chrome's own password manager steals focus from the VK form.
		Prefs: map[string]any{
			"credentials_enable_service": false,
			"password_manager_enabled":   false,
		},
	})
	return caps
}

func sendKeys(wd selenium.WebDriver, xpath, text, what string) error {
	el, err := wd.FindElement(selenium.ByXPATH, xpath)
	if err != nil {
		return fmt.Errorf("find %s: %w", what, err)
	}
	if err := el.SendKeys(text); err != nil {
		return fmt.Errorf("fill %s: %w", what, err)
	}
	return nil
}

func click(wd selenium.WebDriver, xpath, what string) error {
	el, err := wd.FindElement(selenium.ByXPATH, xpath)
	if err != nil {
		return fmt.Errorf("find %s: %w", what, err)
	}
	if err := el.Click(); err != nil {
		return fmt.Errorf("click %s: %w", what, err)
	}
	return nil
}

func screenshot(wd selenium.WebDriver, path string) error {
	data, err := wd.Screenshot()
	if err != nil {
		return fmt.Errorf("capture screenshot: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write screenshot to %s: %w", path, err)
	}
	return nil
}
