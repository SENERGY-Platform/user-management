## Generate File in Version 2.4.0
```
cd asyncapi-gen
go generate ./...
```

## Convert to Version 3.0.0
copy/paste to https://studio.asyncapi.com
a dialog to convert should pop up. if not, you must clear your cookies/local-storage.

alternatively, you could use the `asyncapi convert` cli described in https://www.asyncapi.com/docs/migration/migrating-to-v3

## Additional Annotations
the resulting asyncapi.json should be modified
- channels.Service-Topic.address = "<Service-Topic>"
- channels.Service-Topic.description = "topic is a service.Id with replaced '#' and ':' by '_'"
- channels.Service-Topic.servers = [{"$ref": "#/servers/kafka"}]
- channels.device-types.servers = [{"$ref": "#/servers/kafka"}]
- channels.process-deployment-done.servers = [{"$ref": "#/servers/kafka"}]
- channels.event/{device-local-id}/{service-local-id}.servers = [{"$ref": "#/servers/mqtt"}]
