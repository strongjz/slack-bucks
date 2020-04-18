provider "aws" {
  region                  = "us-west-2"
  profile                 = "strongjz-tech"
  shared_credentials_file = "~/.aws/credentials"
}

provider "aws" {
  alias = "east"
  region                  = "us-east-1"
  profile                 = "strongjz-tech"
  shared_credentials_file = "~/.aws/credentials"
}
