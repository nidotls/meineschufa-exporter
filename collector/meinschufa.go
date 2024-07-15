package collector

import (
	"fmt"
	"github.com/playwright-community/playwright-go"
	"os"
	"strconv"
	"time"
)

func NewSchufaApp(smsApp SmsApp) (SchufaApp, error) {
	return &playrightSchufaApp{
		smsApp:   smsApp,
		username: os.Getenv("MEINESCHUFA_USERNAME"),
		password: os.Getenv("MEINESCHUFA_PASSWORD"),
	}, nil
}

type SchufaResponseEntity struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Details  string `json:"details"`
	Date     string `json:"date"`
}

type SchufaResponse struct {
	Score float32 `json:"score"`
	//RequestDate string                 `json:"requestDate"`
	Datalist []SchufaResponseEntity `json:"datalist"`
}

type SchufaApp interface {
	GetScore() (*SchufaResponse, error)
}

type playrightSchufaApp struct {
	smsApp   SmsApp
	username string
	password string
}

func (a *playrightSchufaApp) GetScore() (*SchufaResponse, error) {
	err := playwright.Install()

	if err != nil {
		return nil, err
	}

	pw, err := playwright.Run()

	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %v", err)
	}

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})

	if err != nil {
		return nil, fmt.Errorf("could not launch browser: %v", err)
	}

	page, err := browser.NewPage()

	if err != nil {
		return nil, fmt.Errorf("could not create page: %v", err)
	}

	if _, err = page.Goto("https://pkp.meineschufa.de/sao/uebersicht"); err != nil {
		return nil, fmt.Errorf("could not goto: %v", err)
	}

	cookieAcceptButton := page.GetByText("Alles akzeptieren")
	if cookieAcceptButton != nil {
		err := cookieAcceptButton.Click()

		if err != nil {
			return nil, fmt.Errorf("could not accept cookies: %v", err)
		}
	}

	usernameInput := page.Locator("#username")
	if usernameInput != nil {
		err := page.Locator("#username").Fill(a.username)

		if err != nil {
			return nil, fmt.Errorf("could not enter username: %v", err)
		}

		err = page.Locator("#password").Fill(a.password)

		if err != nil {
			return nil, fmt.Errorf("could not enter password: %v", err)
		}

		err = page.Locator("#password").Press("Enter")

		if err != nil {
			return nil, fmt.Errorf("could not submit login: %v", err)
		}
	}

	sendSmsButton := page.Locator("#sendSms")
	if sendSmsButton != nil {
		err := sendSmsButton.Click()

		if err != nil {
			return nil, fmt.Errorf("could not request 2fa sms: %v", err)
		}

		time.Sleep(10 * time.Second)

		code, err := a.smsApp.GetCode()

		if err != nil {
			return nil, fmt.Errorf("could not get 2fa sms code: %v", err)
		}

		err = page.Locator("#sms-tan").Fill(code)

		if err != nil {
			return nil, fmt.Errorf("could not enter 2fa sms code: %v", err)
		}

		err = page.Locator("#sms-tan").Press("Enter")

		if err != nil {
			return nil, fmt.Errorf("could not submit 2fa sms code: %v", err)
		}
	}

	err = page.Locator("#bonitaet-datum").WaitFor()

	if err != nil {
		return nil, fmt.Errorf("could not load dashboard: %v", err)
	}

	score, err := page.Locator("score-element").GetAttribute("score")

	if err != nil {
		return nil, fmt.Errorf("could not load score: %v", err)
	}

	scoreNumber, err := strconv.ParseFloat(score, 32)

	if err != nil {
		return nil, fmt.Errorf("could not parse score to float: %v", err)
	}

	s := &SchufaResponse{
		Score:    float32(scoreNumber),
		Datalist: []SchufaResponseEntity{},
	}

	entities, err := page.Locator(".sao-kachel").All()

	if err != nil {
		return nil, fmt.Errorf("could not locale .sao-kachel: %v", err)
	}

	for _, entity := range entities {
		result := SchufaResponseEntity{}

		name, err := entity.Locator(".kopfbereich > h3").InnerText()
		if err == nil {
			result.Name = name
		} else {
			return nil, fmt.Errorf("could not locale entity name: %v", err)
		}

		category, err := entity.Locator(".kopfbereich > h3 > small").InnerText()
		if err == nil {
			result.Category = category
		} else {
			return nil, fmt.Errorf("could not locale entity category: %v", err)
		}

		typeLabel, err := entity.Locator(".typ-label").InnerText()
		if err == nil {
			result.Type = typeLabel
		} else {
			return nil, fmt.Errorf("could not locale entity type: %v", err)
		}

		details, err := entity.Locator(".typ > span").InnerText(playwright.LocatorInnerTextOptions{
			Timeout: playwright.Float(1000),
		})
		if err == nil {
			result.Details = details
		} else {
			// ignore error, field is optional
			//return nil, fmt.Errorf("could not locale entity details: %v", err)
		}

		date, err := entity.Locator(".datum").InnerText()
		if err == nil {
			result.Date = date
		} else {
			return nil, fmt.Errorf("could not locale entity date: %v", err)
		}

		s.Datalist = append(s.Datalist, result)
	}

	return s, nil
}
