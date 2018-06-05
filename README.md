islenauts
============

## Deploy with CircleCI

Set following environment variables on CircleCI project settings

- `GCP_PROJECT_ID` - project id of Google Cloud Platform
- `GCP_SECRET_KEY` - base64 encoded JSON key which can deploy to Google App Engine
- `TWITTER_ACCESS_TOKEN` -
- `TWITTER_ACCESS_TOKEN_SECRET` -

Then you can deploy with `git push origin master` and access https://islenauts-dot-GCP_PROJECT_ID.appspot.com/
