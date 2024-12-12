package config

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"

	"github.com/SixtyAI/cli-o-mat/util"
)

type Omat struct {
	Credentials *CredentialCache `yaml:"-"`

	AccountName        string `yaml:"-"`
	OrganizationPrefix string `yaml:"-"`
	Region             string `yaml:"region"`
	ParamPrefix        string `yaml:"-"`
}

type accountInfoConfig struct {
	AccountID   string `json:"account_id"` // nolint: tagliatelle
	Environment string `json:"environment"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Purpose     string `json:"purpose"`
	Slug        string `json:"slug"`
}

func NewOmat(accountName string) *Omat {
	return &Omat{
		AccountName:        accountName,
		OrganizationPrefix: "",
		Region:             "us-east-1",
		ParamPrefix:        "",
	}
}

func (omat *Omat) loadConfigFromEnv() {
	if region, wasSet := os.LookupEnv("OMAT_REGION"); wasSet {
		omat.Region = region
	}
}

func (omat *Omat) LoadConfig() {
	omat.loadConfigFromEnv()
	omat.InitCredentials()

	omat.FetchOrgPrefix()
	omat.FetchAccountInfo()
}

func (omat *Omat) FetchOrgPrefix() {
	ssmClient := ssm.New(omat.Credentials.RootSession, omat.Credentials.RootAWSConfig)
	roleParamName := "/omat/organization_prefix"

	roleParam, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(roleParamName),
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "ParameterNotFound") {
			util.Fatalf(1, "Couldn't find org prefix parameter: %s\n", roleParamName)
		}

		util.Fatalf(1,
			"Error looking up org prefix parameter %s, got: %s\n", roleParamName, err.Error())
	}

	orgPrefix := aws.StringValue(roleParam.Parameter.Value)
	if orgPrefix == "" {
		util.Fatalf(1, "Paramater '%s' was empty.\n", roleParamName)
	}

	omat.OrganizationPrefix = orgPrefix
}

func (omat *Omat) FetchAccountInfo() {
	ssmClient := ssm.New(omat.Credentials.RootSession, omat.Credentials.RootAWSConfig)
	infoParamName := "/omat/account_registry/" + omat.AccountName

	infoParam, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name: aws.String(infoParamName),
	})
	if err != nil {
		if strings.HasPrefix(err.Error(), "ParameterNotFound") {
			util.Fatalf(1, "Couldn't find account info parameter: %s\n", infoParamName)
		}

		util.Fatalf(1,
			"Error looking up account info parameter %s, got: %s\n", infoParamName, err.Error())
	}

	accountInfo := aws.StringValue(infoParam.Parameter.Value)
	if accountInfo == "" {
		util.Fatalf(1, "Paramater '%s' was empty.\n", infoParamName)
	}

	var data accountInfoConfig
	if err = json.Unmarshal([]byte(accountInfo), &data); err != nil {
		util.Fatalf(1, "Couldn't parse account info parameter: %s\nGot: %s\n", infoParamName, accountInfo)
	}

	omat.ParamPrefix = data.Prefix
}

func (omat *Omat) InitCredentials() {
	omat.Credentials = newCredentialCache(omat)
}
