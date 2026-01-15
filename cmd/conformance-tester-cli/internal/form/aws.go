package form

// getAWSSecretFields returns the secret credential fields for AWS provider
func (fd *FormData) getAWSSecretFields() []SecretField {
	return []SecretField{
		{
			Name:     "KKPDatacenter",
			Label:    "AWS KKPDatacenter",
			Value:    &fd.Secrets.AWS.KKPDatacenter,
			Required: true,
		},
		{
			Name:     "AccessKeyID",
			Label:    "AWS AccessKeyID",
			Value:    &fd.Secrets.AWS.AccessKeyID,
			Required: true,
		},
		{
			Name:     "SecretAccessKey",
			Label:    "AWS SecretAccessKey",
			Value:    &fd.Secrets.AWS.SecretAccessKey,
			Required: true,
		},
	}
}
