package v1

type AssumeRoleInput struct {
	RoleArn         string `json:"roleArn" binding:"required" form:"roleArn"`
	SessionDuration int32  `json:"sessionDuration" binding:"required,numeric,min=900,max=43200" form:"sessionDuration"`
}

type GetConsoleUrlInput struct {
	AccountId  string     `json:"accountId" binding:"required,numeric,len=12" form:"accountId"`
	AccessType AccessType `json:"accessType" binding:"required" form:"accessType"`
	Duration   int        `json:"duration" binding:"required,numeric,min=900,max=43200" form:"duration"`
}
