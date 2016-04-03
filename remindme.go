package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/janekolszak/go-pebble"
)

func newID() string {
	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		panic(err)
	}
	return hex.EncodeToString(random)
}

func NewReminder(title, body, delay string, notifyCreate bool) error {
	now := time.Now()
	var eventTime time.Time
	if strings.Contains(delay, "T") {
		t, err := time.Parse("2006-01-02T15:04", delay)
		if err != nil {
			return fmt.Errorf("format should be 2006-01-02T15:04, %v", err)
		}
		eventTime = t
	} else {
		tdelay, err := time.ParseDuration(delay)
		if err != nil {
			return fmt.Errorf("failed to parse delay: %v", err)
		}
		eventTime = now.Add(tdelay)
	}
	layout := &pebble.Layout{
		Type:     "genericPin",
		Title:    fmt.Sprintf("Reminder: %s", title),
		TinyIcon: "system://images/NOTIFICATION_FLAG",
		Body:     body,
	}
	var creationNotification *pebble.Notification
	if notifyCreate {
		creationNotification = &pebble.Notification{
			Layout: &pebble.Layout{
				Type:     "genericPin",
				Title:    "Creation Title",
				TinyIcon: "system://images/NOTIFICATION_FLAG",
				Body:     "Creation Body",
			},
		}
	}

	// updateLayout := pebble.Layout{
	// 	Type:     "genericPin",
	// 	Title:    "Update Title",
	// 	TinyIcon: "system://images/NOTIFICATION_FLAG",
	// 	Body:     "Update Body",
	// }

	// updateNotification := pebble.Notification{
	// 	Layout: &updateLayout,
	// 	Time:   time.Now().Format(time.RFC3339),
	// }

	reminders := pebble.Reminders{}

	reminders = append(reminders, pebble.Reminder{
		Time: eventTime.Add(-10 * time.Second).Format(time.RFC3339),
		Layout: &pebble.Layout{
			Type:     "genericReminder",
			Title:    title,
			TinyIcon: "system://images/NOTIFICATION_FLAG",
		},
	})
	rdelay := eventTime.Sub(now)
	if rdelay > 1*time.Hour {
		reminders = append(reminders, pebble.Reminder{
			Time: eventTime.Add(-15 * time.Minute).Format(time.RFC3339),
			Layout: &pebble.Layout{
				Type:     "genericReminder",
				Title:    title,
				TinyIcon: "system://images/NOTIFICATION_FLAG",
			},
		})
	}
	if rdelay > 2*time.Hour {
		reminders = append(reminders, pebble.Reminder{
			Time: eventTime.Add(-1 * time.Hour).Format(time.RFC3339),
			Layout: &pebble.Layout{
				Type:     "genericReminder",
				Title:    title,
				TinyIcon: "system://images/NOTIFICATION_FLAG",
			},
		})
	}

	pin := pebble.Pin{
		Id:                 newID(),
		Time:               eventTime.Format(time.RFC3339),
		Layout:             layout,
		CreateNotification: creationNotification,
		// UpdateNotification: &updateNotification,
		Reminders: &reminders,
	}

	userPin := pebble.UserPin{
		Pin:   pin,
		Token: os.Getenv("PEBBLE_TOKEN"),
	}
	if userPin.Token == "" {
		return fmt.Errorf("Empty PEBBLE_TOKEN")
	}

	fmt.Println(pin.String())
	// fmt.Printf("r=%v", userPin.Put(http.DefaultClient))
	return userPin.Put(http.DefaultClient)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: blob [options] <title>]")
	flag.PrintDefaults()
	os.Exit(2)
}

func tempFilename(prefix, suffix string) string {
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	return filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(2)
	}
	delay := os.Args[1]
	title := strings.Join(os.Args[2:], " ")
	fpath := tempFilename("remindme_body_", "")
	if err := ioutil.WriteFile(fpath, []byte{}, 0644); err != nil {
		panic(fmt.Sprintf("failed to create temp file: %s", err))
	}
	defer os.Remove(fpath)
	cmd := exec.Command("vim", fpath)
	// Hook vim to the current session
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("failed to start vim: %s", err))
	}
	if err := cmd.Wait(); err != nil {
		panic(fmt.Sprintf("failed to edit: %s", err))
	}
	data, err := ioutil.ReadFile(fpath)
	if err != nil {
		panic(fmt.Sprintf("failed to open temp file: %s", err))
	}
	if err := NewReminder(title, string(data), delay, false); err != nil {
		panic(err)
	}
}
