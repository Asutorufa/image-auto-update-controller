variable "name" {
  type    = string
  default = "image-updater"
}

variable "namespace" {
  type    = string
  default = "image-updater"
}

variable "cri-socket" {
  type    = string
  default = "/run/k0s/containerd.sock"
}
