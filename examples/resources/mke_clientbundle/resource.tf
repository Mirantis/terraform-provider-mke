# Create an MKE client bundle for docker/kubernetes API access
# (uses the MKE user for the provider auth)
resource "mke_clientbundle" "example" {
  label = "my terraform client bundle for admin user"
}

# OPTIONAL: Output the kube yaml of the MKE cluster
output "kubeconfig" {
  sensitive   = true
  description = "the contents of the kubeconfig yaml file"
  value       = mke_clientbundle.example.kube_yaml
}

# OPTIONAL: Create local_file out of the kube config
resource "local_file" "kubeconfig" {
  filename = "kubeconfig.yaml"
  content  = mke_clientbundle.example.kube_yaml
}
