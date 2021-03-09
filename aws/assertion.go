package aws

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

type STSSession struct {
	RoleARN         string    `json:"ROLE_ARN"`
	AccessKeyID     string    `json:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey string    `json:"AWS_SECRET_ACCESS_KEY"`
	SessionToken    string    `json:"AWS_SESSION_TOKEN"`
	Expiration      time.Time `json:"AWS_SESSION_EXPIRATION"`
	Provider        string    `json:"STS_PROVIDER"`
	Issuer          string    `json:"STS_ISSUER"`
	Region          string    `json:"-"`
}

func (s *STSSession) Expired() bool {
	// 5 seconds of fuzz
	if s.Expiration.Before(time.Now().Add(time.Second * 5)) {
		return true
	}
	return false
}

// get list of role ARNs in a SAML Assertion
func GetRoles(assertion string) ([]string, error) {
	roles := []string{}
	reader := strings.NewReader(assertion)
	q, err := xmlquery.Parse(reader)
	if err != nil {
		return roles, err
	}
	// returns our Roles as well as our Email address
	items := xmlquery.Find(q, "/samlp:Response/saml:Assertion/saml:AttributeStatement/saml:Attribute/saml:AttributeValue")
	for _, item := range items {
		if strings.HasPrefix(item.InnerText(), "arn:aws:iam:") {
			splits := strings.Split(item.InnerText(), ",")
			roles = append(roles, splits[0])
		}
	}
	return roles, nil
}

// get the ARN of the provided role ARN from the saml assertion
func GetRolePrincipalARN(assertion string, role string) (string, error) {
	reader := strings.NewReader(assertion)
	q, err := xmlquery.Parse(reader)
	if err != nil {
		return "", err
	}
	// returns our Roles as well as our Email address
	items := xmlquery.Find(q, "/samlp:Response/saml:Assertion/saml:AttributeStatement/saml:Attribute/saml:AttributeValue")
	search_role := fmt.Sprintf("%s,", role)
	for _, item := range items {
		if strings.HasPrefix(item.InnerText(), search_role) {
			splits := strings.Split(item.InnerText(), ",")
			return splits[1], nil
		}
	}
	return "", fmt.Errorf("Unable to find the role '%s' in SAML assertion", role)
}

// When does this SAML expire?
func GetExpireTime(assertion string) (*time.Time, error) {
	reader := strings.NewReader(assertion)
	q, err := xmlquery.Parse(reader)
	if err != nil {
		return nil, err
	}
	var t time.Time
	// returns our expire time condition
	items := xmlquery.Find(q, "/samlp:Response/saml:Assertion/saml:Conditions/@NotOnOrAfter")
	for _, item := range items {
		t, err = time.Parse("2006-01-02T15:04:05Z", item.InnerText())
		if err != nil {
			return nil, fmt.Errorf("Unable to parse NotOnOrAfter %s: %s", t, err.Error())
		}
		return &t, nil
	}
	return nil, fmt.Errorf("Unable to locate NotOnOrAfter time in SAML Assertion")
}

func GetSTSSession(assertion string, role string, region string, duration int64) (STSSession, error) {
	ret := STSSession{}
	principal, err := GetRolePrincipalARN(assertion, role)
	if err != nil {
		return ret, err
	}
	saml := base64.StdEncoding.EncodeToString([]byte(assertion))

	s, err := session.NewSession()
	if err != nil {
		return ret, err
	}
	svc := sts.New(s, aws.NewConfig().WithRegion(region))
	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: &duration,
		PrincipalArn:    &principal,
		RoleArn:         &role,
		SAMLAssertion:   &saml,
	}
	output, err := svc.AssumeRoleWithSAML(&input)
	if err != nil {
		return ret, err
	}

	creds := output.Credentials

	ret = STSSession{
		RoleARN:         role,
		AccessKeyID:     *creds.AccessKeyId,
		SecretAccessKey: *creds.SecretAccessKey,
		SessionToken:    *creds.SessionToken,
		Expiration:      *creds.Expiration,
		Issuer:          *output.Issuer,
		Region:          region,
	}
	log.Debugf("STSSession = %s", spew.Sdump(ret))
	return ret, nil
}
