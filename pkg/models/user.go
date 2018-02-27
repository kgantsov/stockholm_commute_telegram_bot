package models

type User struct {
	ID       int    `bson:"id"`
	Name     string `bson:"name"`
	ChatID   int64  `bson:"chat_id"`
	HomeName string `bson:"home_name"`
	HomeID   string `bson:"home_id"`
	HomeTime string `bson:"home_time"`
	WorkName string `bson:"work_name"`
	WorkID   string `bson:"work_id"`
	WorkTime string `bson:"work_time"`
}
