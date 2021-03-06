package ec2

import (
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/mitchellh/mapstructure"
	"github.com/cloudquery/cloudquery/providers/common"
	"go.uber.org/zap"
)

type CustomerGateway struct {
	ID                uint `gorm:"primarykey"`
	AccountID         string
	Region            string
	BgpAsn            *string
	CertificateArn    *string
	CustomerGatewayId *string
	DeviceName        *string
	IpAddress         *string
	State             *string
	Tags              []*CustomerGatewayTag `gorm:"constraint:OnDelete:CASCADE;"`
	Type              *string
}

type CustomerGatewayTag struct {
	ID                uint `gorm:"primarykey"`
	CustomerGatewayID uint
	Key               *string
	Value             *string
}

func (c *Client) transformCustomerGatewayTag(value *ec2.Tag) *CustomerGatewayTag {
	return &CustomerGatewayTag{
		Key:   value.Key,
		Value: value.Value,
	}
}

func (c *Client) transformCustomerGatewayTags(values []*ec2.Tag) []*CustomerGatewayTag {
	var tValues []*CustomerGatewayTag
	for _, v := range values {
		tValues = append(tValues, c.transformCustomerGatewayTag(v))
	}
	return tValues
}

func (c *Client) transformCustomerGateway(value *ec2.CustomerGateway) *CustomerGateway {
	return &CustomerGateway{
		Region:            c.region,
		AccountID:         c.accountID,
		BgpAsn:            value.BgpAsn,
		CertificateArn:    value.CertificateArn,
		CustomerGatewayId: value.CustomerGatewayId,
		DeviceName:        value.DeviceName,
		IpAddress:         value.IpAddress,
		State:             value.State,
		Tags:              c.transformCustomerGatewayTags(value.Tags),
		Type:              value.Type,
	}
}

func (c *Client) transformCustomerGateways(values []*ec2.CustomerGateway) []*CustomerGateway {
	var tValues []*CustomerGateway
	for _, v := range values {
		tValues = append(tValues, c.transformCustomerGateway(v))
	}
	return tValues
}

func (c *Client) CustomerGateways(gConfig interface{}) error {
	var config ec2.DescribeCustomerGatewaysInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}
	if !c.resourceMigrated["ec2CustomerGateway"] {
		err := c.db.AutoMigrate(
			&CustomerGateway{},
			&CustomerGatewayTag{},
		)
		if err != nil {
			return err
		}
		c.resourceMigrated["ec2CustomerGateway"] = true
	}

	output, err := c.svc.DescribeCustomerGateways(&config)
	if err != nil {
		return err
	}
	c.log.Debug("deleting previous CustomerGateways", zap.String("region", c.region), zap.String("account_id", c.accountID))
	c.db.Where("region = ?", c.region).Where("account_id = ?", c.accountID).Delete(&CustomerGateway{})
	common.ChunkedCreate(c.db, c.transformCustomerGateways(output.CustomerGateways))
	c.log.Info("populating CustomerGateways", zap.Int("count", len(output.CustomerGateways)))
	return nil
}
