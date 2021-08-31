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
		log.Println("ERROR: DeleteWaitingRoomUser()", err)
		return err
	}
	err = DeleteDashboardUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteDashboardUser()", err)
		return err
	}
	err = DeleteProcessSchedulerUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteProcessSchedulerUser()", err)
		return err
	}
	err = DeleteImportsUser(token, conf)
	if err != nil {
		log.Println("ERROR: DeleteProcessSchedulerUser()", err)
		return err
	}
	err = DeleteKeycloakUser(userId, conf)
	if err != nil {
		log.Println("ERROR: DeleteKeycloakUser()", err)
		return err
	}
	return nil
}

type IdWrapper struct {
	Id string `json:"id"`
}

type WaitingRoomListIdWrapper struct {
	Result []WaitingRoomIdWrapper `json:"result"`
}

type WaitingRoomIdWrapper struct {
	Id string `json:"local_id"`
}
