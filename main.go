package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/markcheno/go-quote"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

func main() {
	api := slack.New("xoxb-4513238464769-4525281164720-aE3YgUJCzNfSEuAxbi6UHya5")

	http.HandleFunc("/slack/events", func(w http.ResponseWriter, r *http.Request) {
		log.Println("start API")
		verifier, err := slack.NewSecretsVerifier(r.Header, "3b738e70c83b66336db5bacf4432bb6d")
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bodyReader := io.TeeReader(r.Body, &verifier)
		body, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := verifier.Ensure(); err != nil {
			log.Println("verify err: ", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		switch eventsAPIEvent.Type {
		case slackevents.URLVerification:
			var res *slackevents.ChallengeResponse
			if err := json.Unmarshal(body, &res); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/plain")
			if _, err := w.Write([]byte(res.Challenge)); err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		case slackevents.CallbackEvent:
			innerEvent := eventsAPIEvent.InnerEvent
			switch event := innerEvent.Data.(type) {
			case *slackevents.AppMentionEvent:
				message := strings.Split(event.Text, " ")
				if len(message) < 2 {
					w.WriteHeader(http.StatusBadRequest)
					return
				}

                yesterday := time.Now().Add(-24 * time.Hour)
                today := time.Now()
                
                start_day := strings.Split(yesterday.String(), " ")[0]
                end_day := strings.Split(today.String(), " ")[0]
                

				brand := message[1]
                span := "d"

                if len(message) >= 3 {
                    start_day = message[2]
                }
                if len(message) >= 4 {
                    end_day = message[3]
                }
                if len(message) >= 5 {
                    span = message[4]
                }


				fmt.Println(brand)
                fmt.Println(start_day)
                fmt.Println(end_day)
                fmt.Println(span)

				spy, err := quote.NewQuoteFromYahoo(brand, start_day, end_day, quote.Period(span), true)
                if err != nil {
                    log.Println(err)
                    return 
                }
				fmt.Print(spy.CSV())
				// rsi2 := talib.Rsi(spy.Close, 2)
				// fmt.Println(rsi2)

				if _, _, err := api.PostMessage(event.Channel, slack.MsgOptionText(spy.CSV(), false)); err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

			}
		}
	})

	log.Println("[INFO] Server listening port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
