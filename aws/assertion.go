package aws

import (
	"fmt"
	"strings"
	"time"

	"github.com/antchfx/xmlquery"
)

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
