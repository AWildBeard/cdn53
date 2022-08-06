# cdn53
_A tool and library that allows you to store arbitrary files in DNS TXT records. Useful for Malware, in-memory execution droppers, etc._

## Features
* Cross-platform golang lib to download files from DNS TXT records created by this tool
* Allows piping a "file" into DNS TXT records on AWS route53
* Allows deleting "file"s from route53 DNS TXT record sets
* Utility download command for downloading files from DNS TXT records created by this command

## Specifying AWS API Credentials
The credential chain looks for credentials in the following order:

1) Environment variables.
   * Static Credentials (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_SESSION_TOKEN`)
   * Web Identity Token (`AWS_WEB_IDENTITY_TOKEN_FILE`)
2) Shared configuration files. These are for boxes that have the `aws` command installed and configured.
   * `credentials` file under `.aws` folder that is in the users home folder.
   * `config` file under `.aws` folder that is in the users home folder.
3) IAM role for tasks. If running as a 'Task'.
   * Yea IDK what this is.
4) When running on an Amazon EC2 instance; IAM role for Amazon EC2.
   * In this scenario, simply having the binary on the EC2 instance will allow the binary to automagically retrieve
   the appropriate access credentials provided the EC2 instance is assigned appropriate privileges through IAM roles.
     
## Getting a binary
1) Clone
   * `git clone git@github.com:AWildBeard/cdn53.git && cd cdn53`
2) Compile for your platform
   * `make release build`
3) Binary will be a executable located in `$PWD/build`

## Usage
```text
This tooling has the ability to;
1) Transfer a local file with encoding into DNS TXT records on route53
2) Delete those TXT records from route53 (cleanup)
3) Download the file and decode it via DNS

Usage:
  cdn53 [command]

Available Commands:
  aws         Create or Delete "cdn"s on route53.
  completion  Generate the autocompletion script for the specified shell
  download    Download the file from the provided "cdn"
  help        Help about any command

Flags:
  -d, --domain [create delete download]   The domain to use for various options ([create delete download]). With the `create` subcommand, this is the FQDN to create a "cdn" with. With the `delete` subcommand, this is the FQDN to delete a "cdn". With the `download` subcommand, this is the FQDN of the "cdn" just as you would've supplied it to the other subcommands. If you created the "cdn" with x.y.z.com then the value should be x.y.z.com
  -f, --file string                       File to read or write content to/from depending on subcommand. If using the create subcommand, this file is read from to create the dns cdn after applying encoding. If using the download subcommand, this file is written to with decoding applied. (default "-")
  -h, --help                              help for cdn53

Use "cdn53 [command] --help" for more information about a command.
```

## Examples
- `cdn53 download -d my.domain.com`
  - Download the file stored in `my.domain.com`'s TXT record set by this tool
- `cat file.txt | cdn53 aws create -d my.domain.com`
  - Store file.txt in TXT records for my.domain.com
- `cdn53 aws create -d my.domain.com -f file.txt`
  - Same as above
- `cdn53 aws delete -d my.domain.com`
  - Delete all the TXT records associated with my.domain.com that were create by this tool to store a file

### Author(s)
- Michael Mitchell

### ToDo:
- [ ] Support base85 encoding (actual base85, not ascii85). Allows for slightly better data usage on TXT records.
