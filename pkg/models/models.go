package models

type User struct {
	ID       int    `bson:"id"`
	Name     string `bson:"name"`
	HomeName string `bson:"home_name"`
	HomeID   string `bson:"home_id"`
	WorkName string `bson:"work_name"`
	WorkID   string `bson:"work_id"`
}
