# kafkacat-hk
Kafkacat wrapper for Heroku Kafka. Takes Heroku env vars and runs [kafkacat](https://github.com/edenhill/kafkacat) with the SSL settings configured.

Currently this only works on Linux because of how I pass the cert data through explicit file descriptors.

## Building
```
git clone https://github.com/ctrlaltdel121/kafkacat-hk
go build
```

## Running
Once the correct environment variables are set (see below), the script accepts the same arguments as kafkacat and passes them through.


## Environment Variables
### Required
`KAFKA_TRUSTED_CERT` - The CA cert that signed the Kafka server's cert

`KAFKA_CLIENT_CERT` - The cert for your Kafka client

`KAFKA_CLIENT_CERT_KEY` - key for KAFKA_CLIENT_CERT


### Optional
`KAFKA_URL` - comma separated list of URLS. Any `://` prefix will be removed. You can also leave this blank and use `-b` option.

`HEROKU` - set this to a non-empty value if your ENV vars are formatted the same as Heroku. If blank, this script assumes base64 encoded environment variables.

`KAFKACAT_BIN` - where your kafkacat binary lives. Defaults to /usr/bin/kafkacat
