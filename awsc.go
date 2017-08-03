package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

var roleArnRegex = regexp.MustCompile(`arn:aws:iam::(?P<accountid>[^:]+):role/(?P<rolename>[^:]+)`)
var roleSessionNameRegex = regexp.MustCompile(`[a-zA-Z0-9+=,.@-]{2,64}`)

type roleInfo struct {
	label string
	arn   string
}

func main() {

	var mfaToken = flag.String("mfa", "", "Your MFA code")
	var duration = flag.Int64("duration", 3600, "Lifetime of credentials.")
	var mfaSerialNumber = flag.String("mfa-serial-number", "", "Your MFA arn")
	var roleArn = flag.String("role", "", "The role name or role ARN")
	var roleSessionName = flag.String("role-session-name", "", "The role session name. Must match [a-zA-Z0-9+=,.@-]{2,64}")
	flag.Parse()
	accountLabel := ""

	var availableRoles []string
	var roleChoices []roleInfo
	availableRolesStr := os.Getenv("AWS_CREDS_ROLES")
	if availableRolesStr != "" {
		availableRoles = strings.Split(availableRolesStr, ",")
		for _, val := range availableRoles {
			maybeRoleLabel := strings.Split(val, ";")
			var role roleInfo
			if len(maybeRoleLabel) == 2 {
				role.label = maybeRoleLabel[0]
				role.arn = maybeRoleLabel[1]
			} else {
				role.label = val
				role.arn = val
			}
			roleChoices = append(roleChoices, role)
		}
	}
	if *roleArn == "" {
		if len(availableRoles) > 0 {
			var choice int
			for {
				fmt.Fprintln(os.Stderr, "Select a role:")
				for idx, val := range roleChoices {
					fmt.Fprintf(os.Stderr, "  %d: %s\n", idx+1, val.label)
				}
				_, err := fmt.Scanf("%d\n", &choice)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
				}
				if choice < 0 || choice > len(availableRoles)+1 {
					fmt.Fprintln(os.Stderr, "Invalid selection")
				}
				break
			}
			*roleArn = roleChoices[choice-1].arn
			accountLabel = roleChoices[choice-1].label
		}
	}
	if *roleArn == "" {
		fmt.Fprintf(os.Stderr, "--role is required.")
	}
	regexResults := roleArnRegex.FindStringSubmatch(*roleArn)
	if regexResults == nil {
		fmt.Fprintf(os.Stderr, "Role ARN doesn't look like an ARN?")
		os.Exit(1)
	}
	roleAccountId := regexResults[1]
	roleName := regexResults[2]
	if accountLabel == "" {
		accountLabel = roleAccountId + "/" + roleName
	}

	if *mfaSerialNumber == "" {
		*mfaSerialNumber = os.Getenv("AWS_MFA_ARN")
		if *mfaSerialNumber == "" {
			fmt.Fprintf(os.Stderr, "Unable to find your MFA ARN. You can specify --mfa-serial-number or $AWS_MFA_ARN. Your ARN is typically of the form 'arn:aws:iam::<account id>:mfa/<username>'")
		}
	}

	if *roleSessionName == "" {
		fmt.Fprintf(os.Stderr, "Enter a role session name (must match [a-zA-Z0-9+=,.@-]{2,64}):\n")
		fmt.Fprintf(os.Stderr, "This is 64 characters long:\n")
		fmt.Fprintf(os.Stderr, "----------------------------------------------------------------\n")
		_, err := fmt.Scanf("%s\n", roleSessionName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid role session name: %s\n", err)
			os.Exit(1)
		}
	}

	if !roleSessionNameRegex.MatchString(*roleSessionName) {
		fmt.Fprintf(os.Stderr, "role-session-name doesn't look valid: '%s'\n", *roleSessionName)
		os.Exit(1)
	}

	if *mfaToken == "" {
		fmt.Fprintln(os.Stderr, "Enter your MFA token for", *mfaSerialNumber)
		_, err := fmt.Scanf("%s\n", mfaToken)
		if err != nil || len(*mfaToken) > 6 {
			fmt.Fprintln(os.Stderr, "Bad MFA token.", err)
			os.Exit(1)
		}
	}

	svc := sts.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
	output, err := svc.AssumeRole(&sts.AssumeRoleInput{
		DurationSeconds: duration,
		RoleArn:         roleArn,
		RoleSessionName: roleSessionName,
		SerialNumber:    mfaSerialNumber,
		TokenCode:       mfaToken,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Unable to get new creds: ", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, `
export AWS_ACCESS_KEY_ID=%s
export AWS_SECRET_ACCESS_KEY=%s
export AWS_SESSION_TOKEN=%s
export AWS_ACCOUNT_ID=%s
export AWS_KEY_NAME=%s
`,
		*output.Credentials.AccessKeyId,
		*output.Credentials.SecretAccessKey,
		*output.Credentials.SessionToken,
		roleAccountId,
		accountLabel,
	)
}
