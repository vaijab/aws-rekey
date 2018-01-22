package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/go-ini/ini"
)

var (
	// Version is set at compile time, passing -ldflags "-X main.Version=<build version>"
	Version  string
	logFatal *log.Logger
	logInfo  *log.Logger
)

func init() {
	logFatal = log.New(os.Stderr, "[error] ", log.Lshortfile)
	logInfo = log.New(os.Stderr, "[info] ", log.Lshortfile)
}

func main() {
	version := flag.Bool("version", false, "Show version and exit")
	credentialsFile := flag.String("credentials-file", "", "The aws credentials file path")
	profile := flag.String("profiles", "default", "The aws credentials profile name(s), comma separated")
	flag.Parse()

	if *version {
		fmt.Printf("aws-rekey %s\n", Version)
		os.Exit(0)
	}

	if len(*credentialsFile) == 0 {
		f, err := lookupCredsFile()
		if err != nil {
			logFatal.Fatal(err)
		}
		*credentialsFile = f
	}

	// Load credentials file and fail hard if an error occurs
	cfg, err := loadCredsFile(*credentialsFile)
	if err != nil {
		logFatal.Fatal(err)
	}

	// Split comma separated list of profile names into profiles slice and trim
	// any leading/trailing commas
	profiles := strings.Split(strings.Trim(*profile, ","), ",")

	for _, p := range profiles {
		// Create a new session for profile p
		sess := session.New(&aws.Config{Credentials: credentials.NewSharedCredentials(*credentialsFile, p)})
		oldCreds, err := sess.Config.Credentials.Get()
		if err != nil {
			logFatal.Fatalf("[%s] failed to read existing credentials: %s.\n", p, err)
		}

		client := iam.New(sess)
		user, err := client.GetUser(&iam.GetUserInput{})
		if err != nil {
			logFatal.Fatalf("[%s] failed to get the IAM user name: %s.\n", p, err)
		}

		// Create a new access key and fail hard if an error occurs
		newCreds, err := newAccessKey(client, user.User.UserName)
		if err != nil {
			logFatal.Fatalf("[%s] failed to create a new access key: %s.\n", p, err)
		}

		// Write the new access key to the credentials file, if an error
		// occurs, print the new key to stdout
		section := cfg.Section(p)
		section.Key("aws_access_key_id").SetValue(*newCreds.AccessKey.AccessKeyId)
		section.Key("aws_secret_access_key").SetValue(*newCreds.AccessKey.SecretAccessKey)
		if err := cfg.SaveTo(*credentialsFile); err != nil {
			logFatal.Printf("[%s] failed to write new access key to %s: %s.\n", p, *credentialsFile, err)
			fmt.Printf("aws_access_key_id: %s\naws_secret_access_key: %s\n",
				*newCreds.AccessKey.AccessKeyId,
				*newCreds.AccessKey.SecretAccessKey,
			)
		}

		logInfo.Printf("[%s] new access key created: %s.\n", p, *newCreds.AccessKey.AccessKeyId)

		if err := deleteAccessKey(client, user.User.UserName, &oldCreds.AccessKeyID); err != nil {
			logFatal.Fatalf("[%s] failed to delete old access key: %s: %s.\n", p, oldCreds.AccessKeyID, err)
		}
	}
}

// lookupCredsFile is a helper which looks up aws shared credentials file in
// the right order and for specific OS.
func lookupCredsFile() (string, error) {
	if f, ok := os.LookupEnv("AWS_SHARED_CREDENTIALS_FILE"); ok {
		return f, nil
	}

	homeDir := os.Getenv("HOME")
	if homeDir == "" { // *nix
		homeDir = os.Getenv("USERPROFILE") // windows
	}
	if homeDir == "" {
		return "", fmt.Errorf("unable to find AWS shared credentials file.\n")
	}

	return filepath.Join(homeDir, ".aws", "credentials"), nil
}

// loadCredsFile loads the credentials file. It will return a *ini.File struct
// and an error if any.
func loadCredsFile(file string) (*ini.File, error) {
	cfg, err := ini.Load(file)
	if err != nil {
		return &ini.File{}, err
	}
	return cfg, nil
}

func newAccessKey(client *iam.IAM, username *string) (*iam.CreateAccessKeyOutput, error) {
	attr := &iam.CreateAccessKeyInput{UserName: username}
	resp, err := client.CreateAccessKey(attr)
	if err != nil {
		return &iam.CreateAccessKeyOutput{}, err
	}
	return resp, nil
}

func deleteAccessKey(client *iam.IAM, username, accessKeyID *string) error {
	attr := &iam.DeleteAccessKeyInput{AccessKeyId: accessKeyID, UserName: username}
	_, err := client.DeleteAccessKey(attr)
	return err
}
