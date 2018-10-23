provides http-api to read and delete users from keycloak.
delete commands will be published on a amqp broker to inform other services about the deletion of the user.