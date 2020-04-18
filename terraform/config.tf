terraform {
  backend "s3" {
    bucket = "acg-advance-network-tf-state"
    key    = "tf.start"
    region = "us-west-2"
  }
}
