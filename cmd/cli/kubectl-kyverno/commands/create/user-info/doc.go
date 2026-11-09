package userinfo

var websiteUrl = `https://kyverno.io/docs/kyverno-cli/#create`

var description = []string{
	`Create a Kyverno user-info file.`,
}

var examples = [][]string{
	{
		"# Create user info file",
		"kyverno create user-info -u molybdenum@somecorp.com -g basic-user -c admin",
	},
	{
		"# Create user info file with custom resource name and output path",
		"kyverno create user-info -n custom-user -u alice -g devs -c admin -o user-info.yaml",
	},
}
