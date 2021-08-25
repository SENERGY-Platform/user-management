package ctrl

import (
	"github.com/SENERGY-Platform/user-management/pkg/configuration"
	"log"
)

func DeleteUser(userId string, conf configuration.Config) (err error) {
	token, err := CreateToken("users-service", userId)
	if err != nil {
		log.Println("ERROR: unable to create jwt for userId", userId, err)
		return err
	}
	err = DeleteWaitingRoomUser(token, conf)
	if err != nil {
		return err
	}
	err = DeleteDashboardUser(token, conf)
	if err != nil {
		return err
	}
	err = DeleteProcessSchedulerUser(token, conf)
	if err != nil {
		return err
	}
	err = DeleteKeycloakUser(userId, conf)
	return err
}

type IdWrapper struct {
	Id string `json:"id"`
}
