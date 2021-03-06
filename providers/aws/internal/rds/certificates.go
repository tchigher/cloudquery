package rds

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/mitchellh/mapstructure"
	"github.com/cloudquery/cloudquery/providers/common"
	"go.uber.org/zap"
	"time"
)

type Certificate struct {
	ID                        uint `gorm:"primarykey"`
	AccountID                 string
	Region                    string
	CertificateArn            *string
	CertificateIdentifier     *string
	CertificateType           *string
	CustomerOverride          *bool
	CustomerOverrideValidTill *time.Time
	Thumbprint                *string
	ValidFrom                 *time.Time
	ValidTill                 *time.Time
}

func (c *Client) transformCertificate(value *rds.Certificate) *Certificate {
	return &Certificate{
		Region:                    c.region,
		AccountID:                 c.accountID,
		CertificateArn:            value.CertificateArn,
		CertificateIdentifier:     value.CertificateIdentifier,
		CertificateType:           value.CertificateType,
		CustomerOverride:          value.CustomerOverride,
		CustomerOverrideValidTill: value.CustomerOverrideValidTill,
		Thumbprint:                value.Thumbprint,
		ValidFrom:                 value.ValidFrom,
		ValidTill:                 value.ValidTill,
	}
}

func (c *Client) transformCertificates(values []*rds.Certificate) []*Certificate {
	var tValues []*Certificate
	for _, v := range values {
		tValues = append(tValues, c.transformCertificate(v))
	}
	return tValues
}

func (c *Client) Certificates(gConfig interface{}) error {
	var config rds.DescribeCertificatesInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}
	if !c.resourceMigrated["rdsCertificate"] {
		err := c.db.AutoMigrate(
			&Certificate{},
		)
		if err != nil {
			return err
		}
		c.resourceMigrated["rdsCertificate"] = true
	}
	for {
		output, err := c.svc.DescribeCertificates(&config)
		if err != nil {
			return err
		}
		c.log.Debug("deleting previous Certificates", zap.String("region", c.region), zap.String("account_id", c.accountID))
		c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&Certificate{})
		common.ChunkedCreate(c.db, c.transformCertificates(output.Certificates))
		c.log.Info("populating Certificates", zap.Int("count", len(output.Certificates)))
		if aws.StringValue(output.Marker) == "" {
			break
		}
		config.Marker = output.Marker
	}
	return nil
}
