# aws-creds
Tool for aiding jumping between AWS accounts

The idea is that you have an IAM user account that has permission to
call AssumeRole, possibly only to specific roles, and MFA is required to
use that user account.

    export AWS_ACCESS_KEY_ID=AK...
    export AWS_SECRET_ACCESS_KEY=...
    export AWS_MFA_ARN=arn:aws:iam::<account id>:mfa/<username>

At this point awsc is almost ready to run. Now you use environment variables to configure which roles you can assume. The syntax for this:

	ROLES => ROLE_OPTION [ COMMA ROLE_OPTION [COMMA ROLE_OPTION]]
    ROLE_OPTION => LABEL SEMICOLON ROLE_ARN
    LABEL => [a-zA-Z0-9_-]+
    ROLE_ARN => <aws role arn>
    SEMICOLON => ;
    COMMA => ,

An example:

		export AWS_CREDS_ROLES="\
		bvz-power;arn:aws:iam::<account id 1>:role/power,\
		bvz-read;arn:aws:iam::<account id 1>:role/read,\
		shared-admin;arn:aws:iam::<other account id>:role/admin,\
        "


Now you're ready to run awsc:

	$ awsc
	Select a role:
	  1: bvz-power
	  2: bvz-read
	  3: shared-admin
	2
	Enter a role session name (must match [a-zA-Z0-9+=,.@-]{2,64}):
	This is 64 characters long:
	----------------------------------------------------------------
	example-time
	Enter your MFA token for arn:aws:iam::123456789012:mfa/bob-cli
	527913

	export AWS_ACCESS_KEY_ID=...
	export AWS_SECRET_ACCESS_KEY=...
	export AWS_SESSION_TOKEN=...
	export AWS_ACCOUNT_ID=<account id 1>
	export AWS_KEY_NAME=bvz-read

The idea is that you could run this as eval $(awsc).

