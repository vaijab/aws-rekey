package main

import (
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/go-ini/ini"
	"log"
	"os"
	"path/filepath"
)

// aws-rekey --profile vaijab
// aws-rekey -c /home/vaijab/.aws/credentials
// aws-rekey --all-profiles
// aws-rekey --daemon

func main() {
	credentialsFile := flag.String("credentials-file", "", "The aws credentials file path")
	profile := flag.String("profile", "default", "The aws credentials profile name")
	// allProfiles := flag.Bool("all-profiles", false, "Re-key all available proviles in the aws credentials file")
	flag.Parse()

	if len(*credentialsFile) == 0 {
		f, err := lookupCredsFile()
		if err != nil {
			log.Fatal(err)
		}
		*credentialsFile = f
	}

	sess := session.New(&aws.Config{Credentials: credentials.NewSharedCredentials(*credentialsFile, *profile)})
	// sess := session.New(&aws.Config{Credentials: credentials.NewCredentials(&credentials.SharedCredentialsProvider{Profile: *profile})})
	creds, err := sess.Config.Credentials.Get()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("AccessKeyID = %+v\nSecretAccessKey = %v\n", creds.AccessKeyID, creds.SecretAccessKey)

	client := iam.New(sess)
	user, err := client.GetUser(&iam.GetUserInput{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("user = %+v\n", *user.User.UserName)

	newCreds, err := createAccessKey(client, user.User.UserName)
	if err != nil {
		log.Fatal(err)
	}

	cfg, err := ini.Load(*credentialsFile)
	if err != nil {
		log.Fatal(err)
	}
	section := cfg.Section(*profile)
	section.Key("aws_access_key_id").SetValue(*newCreds.AccessKey.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(*newCreds.AccessKey.SecretAccessKey)
	if err := cfg.SaveTo(*credentialsFile); err != nil {
		log.Fatal(err)
	}

	err = deleteAccessKey(client, user.User.UserName, &creds.AccessKeyID)
	if err != nil {
		log.Fatal(err)
	}
	// load credentials file

	// create a new access key

	// write the new access key to the credentials file

	// delete the old access key

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
		return "", fmt.Errorf("Unable to find AWS shared credentials file.\n")
	}

	return filepath.Join(homeDir, ".aws", "credentials"), nil
}

// loadCredsFile loads the credentials file. It will return a *ini.File struct.
func loadCredsFile(file string) (*ini.File, error) {

	return &ini.File{}, nil
}

func createAccessKey(client *iam.IAM, username *string) (*iam.CreateAccessKeyOutput, error) {
	attr := &iam.CreateAccessKeyInput{UserName: username}
	resp, err := client.CreateAccessKey(attr)
	if err != nil {
		return &iam.CreateAccessKeyOutput{}, err
	}
	// FIXME: remove below
	fmt.Printf(
		"New access keys:\n- AccessKeyID = %+v\n- SecretAccessKey = %v\n",
		*resp.AccessKey.AccessKeyId,
		*resp.AccessKey.SecretAccessKey,
	)
	return resp, err
}

func deleteAccessKey(client *iam.IAM, username, accessKeyID *string) error {
	attr := &iam.DeleteAccessKeyInput{AccessKeyId: accessKeyID, UserName: username}
	_, err := client.DeleteAccessKey(attr)
	return err
}
