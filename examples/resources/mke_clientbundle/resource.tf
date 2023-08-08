# Create an MKE client bundle for docker/kubernetes API access
# (uses the MKE user for the provider auth)
resource "mke_clientbundle" "example" {
  label = "my terraform client bundle for admin user"
}
