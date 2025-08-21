resource "boxer_policy_set" "vacation_photo_access_policy" {
  id         = "vacation-photo-access-policy"
  schema     = boxer_validator_cedar_schema.integration_test.id
  data_cedar = <<EOT
  permit (
      principal == PhotoApp::User::"alice",
      action == PhotoApp::Action::"viewPhoto",
      resource == PhotoApp::Photo::"vacationPhoto.jpg"
  );
EOT
}
