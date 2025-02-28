provider "kubernetes" {
  config_path = "~/.kube/config"
}

resource "kubernetes_service_account" "image-updater" {
  metadata {
    name      = var.name
    namespace = var.namespace
  }
}

resource "kubernetes_cluster_role" "image-updater" {
  metadata {
    name = var.name
  }

  rule {
    api_groups = [""]
    resources  = ["pods"]
    // create,delete,deletecollection,get,list,patch,update,watch
    verbs = ["get", "list", "watch", "delete", "deletecollection", "update", "patch"]
  }
}

resource "kubernetes_cluster_role_binding" "image-updater" {
  metadata {
    name = var.name
  }

  role_ref {
    api_group = "rbac.authorization.k8s.io"
    kind      = "ClusterRole"
    name      = kubernetes_cluster_role.image-updater.metadata[0].name
  }

  subject {
    kind      = "ServiceAccount"
    name      = kubernetes_service_account.image-updater.metadata[0].name
    namespace = var.namespace
  }
}


resource "kubernetes_daemonset" "image-updater" {
  metadata {
    name      = var.name
    namespace = var.namespace
  }

  spec {
    selector {
      match_labels = {
        app = var.name
      }
    }

    template {
      metadata {
        labels = {
          app = var.name
        }
      }

      spec {
        service_account_name = kubernetes_service_account.image-updater.metadata[0].name
        container {
          name              = var.name
          image             = "ghcr.io/Asutorufa/image-auto-update-controller:main"
          image_pull_policy = "IfNotPresent"

          env {
            name  = "CRI_ENDPOINT_ADDRESS"
            value = var.cri-socket
          }

          env {
            name  = "UPDATE_TICKER"
            value = "24"
          }

          volume_mount {
            name       = "cri-socket"
            mount_path = var.cri-socket
          }
        }

        volume {
          name = "cri-socket"
          host_path {
            path = var.cri-socket
          }
        }
      }
    }
  }
}
