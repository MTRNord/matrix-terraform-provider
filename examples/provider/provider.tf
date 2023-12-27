provider "matrix" {
  # The client/server URL to access your matrix homeserver with.
  # Environment variable: MATRIX_CLIENT_SERVER_URL
  client_server_url = "https://matrix.org"

  # The default access token to use for things like content uploads.
  # Does not apply for provisioning users.
  # Environment variable: MATRIX_DEFAULT_ACCESS_TOKEN
  default_access_token = "MDAxSomeRandomString"

  # The default userID to use for things like content uploads.
  # Must match the user that owns the default_access_token.
  # Does not apply for provisioning users.
  # Environment variable: MATRIX_DEFAULT_USERID
  default_user_id = "@meow:matrix.org"
}