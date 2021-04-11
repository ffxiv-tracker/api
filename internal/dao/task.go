package dao

type UserMasterTaskListKey struct {
	UserID string `json:"PK"`
	Type string `json:"SK"`
}

type UserMasterTaskListUpdate struct {
	Tasks []string `dynamodbav:":t,stringset"`
}

type UserMasterTaskListUpdated struct {
	Tasks []*string `json:"tasks"`
}